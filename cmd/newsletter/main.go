package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/club-1/newsletter-go"
)

const CmdName = "newsletter"

var verbose bool

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

func printPreview(mail *newsletter.Mail) {
	fmt.Print("================ PREVIEW START ================\n")
	fmt.Print("┌---- Header ------\n")
	fmt.Printf("| Subject: %s\n", mail.Subject)
	fmt.Printf("| From: %s\n", mail.From())
	fmt.Print("└------------------\n")
	fmt.Printf("%s\n", mail.Body)
	fmt.Print("================  PREVIEW END  ================\n")
}

func sendPreviewMail(mail *newsletter.Mail) error {
	mail.To = newsletter.LocalUserAddr()
	mail.Subject += " (preview)"

	err := newsletter.SendMail(mail)
	if err != nil {
		return fmt.Errorf("could not send preview mail: %w", err)
	}
	fmt.Printf("📨 preview email send to %s\n", newsletter.LocalUserAddr())
	return nil
}

func initialize() error {
	prefix, err := getCmdPrefix()
	if err != nil {
		return fmt.Errorf("could not get command prefix: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get user home directory: %w", err)
	}

	errCount := 0
	for _, route := range newsletter.Routes {
		fileName := ".forward+" + route
		filePath := filepath.Join(homeDir, fileName)
		if verbose {
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
	if errCount > 0 {
		return fmt.Errorf("could not write %v files", errCount)
	}
	return nil
}

func preview(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("missing arguments")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments")
	}
	subject := args[0]
	bodyPath := args[1]
	bodyB, err := os.ReadFile(bodyPath)
	if err != nil {
		return fmt.Errorf("could not load newsletter body: %w", err)
	}

	mail := newsletter.DefaultMail(subject, string(bodyB))
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", newsletter.UnsubscribeAddr())

	return sendPreviewMail(mail)
}

func send(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("missing arguments")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments")
	}
	subject := args[0]
	bodyPath := args[1]
	bodyB, err := os.ReadFile(bodyPath)
	if err != nil {
		return fmt.Errorf("could not load newsletter body: %w", err)
	}

	mail := newsletter.DefaultMail(subject, string(bodyB))
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", newsletter.UnsubscribeAddr())

	err = sendPreviewMail(mail)
	if err != nil {
		return err
	}

	addrCount := len(newsletter.Conf.Emails)
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
		return fmt.Errorf("could not build confirm form: %w", err)
	}
	if !confirm {
		fmt.Printf("❌ sending aborted\n")
		os.Exit(2)
	}

	fmt.Print("sending")
	var errCount = 0
	for _, address := range newsletter.Conf.Emails {
		mail.To = address
		err := newsletter.SendMail(mail)
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
	return nil
}

func main() {
	logFile := newsletter.InitLogger(CmdName)
	defer logFile.Close()

	err := newsletter.ReadConfig()
	if err != nil {
		log.Printf("init: %v", err)
	}

	flag.BoolVar(&verbose, "v", false, "increase verbosity of program")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		log.Fatalf("missing subcommand")
	}

	var cmdErr error

	switch args[0] {
	case "init":
		cmdErr = initialize()
	case "send":
		cmdErr = send(args[1:])
	case "preview":
		cmdErr = preview(args[1:])
	default:
		log.Fatalln("invalid sub command")
	}

	if cmdErr != nil {
		log.Fatalf("error: %v", cmdErr)
	}
}
