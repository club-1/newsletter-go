// This file is part of club-1/newsletter-go.
//
// Copyright (c) 2026 CLUB1 Members <contact@club1.fr>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/club-1/newsletter-go/v3"
	"github.com/club-1/newsletter-go/v3/mailer"
	"github.com/club-1/newsletter-go/v3/messages"
)

const CmdName = "newsletter"

var (
	nl          *newsletter.Newsletter
	flagVerbose bool
	flagYes     bool
	flagPreview bool
	flagHelp    bool
)

func getCmdPrefix() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}
	realpath, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("eval symlinks: %w", err)
	}
	return filepath.Dir(filepath.Dir(realpath)), nil
}

// return subject and body contents.
// if second argument is ommited, body content is read through STDIN
func getSubjectBody(args []string) (string, string, error) {
	var bodyB []byte
	var err error
	switch len(args) {
	case 0:
		return "", "", fmt.Errorf("missing arguments")

	case 1:
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return "", "", fmt.Errorf("missing STDIN piped input")
		}
		bodyB, err = io.ReadAll(os.Stdin)
		if err != nil {
			return "", "", fmt.Errorf("read content from STDIN: %w", err)
		}

	case 2:
		bodyPath := args[1]
		bodyB, err = os.ReadFile(bodyPath)
		if err != nil {
			return "", "", fmt.Errorf("load newsletter body: %w", err)
		}

	default:
		return "", "", fmt.Errorf("too many arguments")
	}
	return args[0], string(bodyB), nil
}

func printPreview(mail *mailer.Mail) {
	fmt.Print("================ PREVIEW START ================\n")
	fmt.Print("┌---- Header ------\n")
	fmt.Printf("| Subject: %s\n", mail.Subject)
	fmt.Printf("| From: %s\n", mail.From())
	fmt.Print("└------------------\n")
	fmt.Printf("%s\n", mail.Body)
	fmt.Print("================  PREVIEW END  ================\n")
}

func initForwardFiles() error {
	prefix, err := getCmdPrefix()
	if err != nil {
		return fmt.Errorf("get command prefix: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get user home directory: %w", err)
	}

	errCount := 0
	for _, route := range newsletter.Routes {
		fileName := ".forward+" + route
		filePath := filepath.Join(homeDir, fileName)
		_, err = os.Stat(filePath)
		if errors.Is(err, os.ErrNotExist) {
			if flagVerbose {
				fmt.Printf("writting file %q\n", filePath)
			}

			cmdPath := filepath.Join(prefix, "sbin/newsletterctl")
			content := []byte("| \"" + cmdPath + " " + route + "\"\n")
			err := os.WriteFile(filePath, content, 0664)
			if err != nil {
				log.Printf("cannot write file %q: %v", filePath, err)
				errCount++
			}
		}
	}
	if errCount > 0 {
		return fmt.Errorf("write %v file(s)", errCount)
	}
	return nil
}

func stop() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get user home directory: %w", err)
	}

	errCount := 0
	for _, route := range newsletter.Routes {
		fileName := ".forward+" + route
		filePath := filepath.Join(homeDir, fileName)
		if flagVerbose {
			fmt.Printf("deleting file %q\n", filePath)
		}

		err := os.Remove(filePath)
		if err != nil {
			log.Printf("cannot delete file %q: %v", filePath, err)
			errCount++
		}
	}
	if errCount > 0 {
		return fmt.Errorf("remove %v file(s)", errCount)
	}
	return nil
}

func setup() error {
	err := initForwardFiles()
	if err != nil {
		return err
	}

	setupForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Newsletter title ?").
				Description("It will be visible before the subject inside square brackets").
				Value(&nl.Config.Settings.Title),
			huh.NewInput().
				Title("Sender displayed name").
				Description("Newsletter sender's name").
				Value(&nl.Config.Settings.DisplayName),
		),
		huh.NewGroup(
			huh.NewSelect[messages.Language]().
				Title("Language").
				Description("Language used for subscription and unsubscription mails").
				Options(
					huh.NewOption("english", messages.LangEnglish),
					huh.NewOption("french", messages.LangFrench),
				).
				Value(&nl.Config.Settings.Language),
		),
		huh.NewGroup(
			huh.NewText().
				Title("Signature").
				Description("newsletter's signature will be inserted under each newsletter").
				Value(&nl.Config.Signature),
		),
	)
	if err := setupForm.Run(); err != nil {
		return fmt.Errorf("build setup form: %w", err)
	}
	err = nl.Config.SaveSettings()
	if err != nil {
		return err
	}
	if flagVerbose {
		fmt.Printf("settings sucessfully saved to file %q\n", newsletter.SettingsFile)
	}
	err = nl.Config.SaveSignature()
	if err != nil {
		return err
	}
	if flagVerbose {
		fmt.Printf("signature sucessfully saved to file %q\n", newsletter.SignatureFile)
	}
	fmt.Println("💾 saved !")
	return nil
}

func send(args []string) error {
	subject, body, err := getSubjectBody(args)
	if err != nil {
		return err
	}

	mail := nl.DefaultMail(subject, body)
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", nl.UnsubscribeAddr())

	addrCount := len(nl.Config.Emails)

	if !flagYes {
		err = nl.SendPreviewMail(mail)
		if err != nil {
			return err
		}

		if flagPreview {
			os.Exit(0)
		}

		duration := float32(addrCount) / 5.0

		var confirm bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Do you really want to send this to %v email addresses ?\n", addrCount)).
					Description(fmt.Sprintf("this will take %v seconds", duration)).
					Value(&confirm),
			),
		)
		if err := confirmForm.Run(); err != nil {
			return fmt.Errorf("build confirm form: %w", err)
		}
		if !confirm {
			fmt.Printf("❌ sending aborted\n")
			os.Exit(2)
		}
	}

	fmt.Print("sending")
	var errCount = 0
	for _, address := range nl.Config.Emails {
		mail.To = address
		err := mailer.Send(mail)
		if err != nil {
			errCount++
			fmt.Print("x")
		} else {
			fmt.Print("·")
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("done !\n")

	if errCount > 0 {
		return fmt.Errorf("error occured while sending mail to %v addresses", errCount)
	}

	log.Printf("newsletter sent to %v email addresses with %v error(s)", addrCount, errCount)
	return nil
}

const banner = "" +
	"      __    __          __   /   __  _/_  _/_    __    __\n" +
	"    /   ) /___)| /| /  (_ ` /  /___) /    /    /___) /   `\n" +
	"___/___/_(___ _|/_|/__(__)_/__(___ _(_ __(_ __(___ _/____v3.0___"

const usage = `
Usage: newsletter [OPTION]... setup
       newsletter [OPTION]... send SUBJECT [CONTENT_FILE]

Options:`

func help() {
	fmt.Println(banner)
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage()
	os.Exit(0)
}

func cmdlineFatalf(format string, v ...any) {
	log.Printf(format, v...)
	flag.Usage()
	os.Exit(2)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.BoolVar(&flagVerbose, "v", false, "verbose: increase verbosity of program")
	flag.BoolVar(&flagYes, "y", false, "yes: always answer yes when program ask for confirmation")
	flag.BoolVar(&flagPreview, "p", false, "preview: limit to a preview (cannot by used with -y)")
	flag.BoolVar(&flagHelp, "h", false, "shorthand for -help")
	flag.BoolVar(&flagHelp, "help", false, "show help message")
	flag.Parse()

	if flagHelp {
		help()
	}

	log.SetFlags(0) // remove all logger flags (remove timestamp)

	if flagYes && flagPreview {
		cmdlineFatalf("illegal combination: -y and -p connot be used at the same time")
	}

	args := flag.Args()
	if len(args) < 1 {
		help()
	}

	var err error
	nl, err = newsletter.New()
	if err != nil {
		log.Fatalf("init newsletter: %v", err)
	}

	var cmdErr error

	switch args[0] {
	case "stop":
		cmdErr = stop()
	case "setup":
		cmdErr = setup()
	case "send":
		cmdErr = send(args[1:])
	default:
		cmdlineFatalf("invalid sub command: %s", args[0])
	}

	if cmdErr != nil {
		log.Fatalf("%s error: %v", args[0], cmdErr)
	}
}
