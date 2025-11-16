package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/club-1/newsletter-go"
)

const CmdName = "newsletter"

var (
	verbose bool
	yes     bool
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
			return "", "", fmt.Errorf("could not read content from STDIN: %w", err)
		}

	case 2:
		bodyPath := args[1]
		bodyB, err = os.ReadFile(bodyPath)
		if err != nil {
			return "", "", fmt.Errorf("could not load newsletter body: %w", err)
		}

	default:
		return "", "", fmt.Errorf("too many arguments")
	}
	return args[0], string(bodyB), nil
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
		return fmt.Errorf("could not write %v file(s)", errCount)
	}
	return nil
}

func stop() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get user home directory: %w", err)
	}

	errCount := 0
	for _, route := range newsletter.Routes {
		fileName := ".forward+" + route
		filePath := filepath.Join(homeDir, fileName)
		if verbose {
			fmt.Printf("deleting file %q\n", filePath)
		}

		err := os.Remove(filePath)
		if err != nil {
			log.Printf("cannot delete file %q: %v", filePath, err)
			errCount++
		}
	}
	if errCount > 0 {
		return fmt.Errorf("could not remove %v file(s)", errCount)
	}
	return nil
}

func preview(args []string) error {
	subject, body, err := getSubjectBody(args)
	if err != nil {
		return err
	}

	mail := newsletter.DefaultMail(subject, body)
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", newsletter.UnsubscribeAddr())

	return sendPreviewMail(mail)
}

func send(args []string) error {
	subject, body, err := getSubjectBody(args)
	if err != nil {
		return err
	}

	mail := newsletter.DefaultMail(subject, body)
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", newsletter.UnsubscribeAddr())

	addrCount := len(newsletter.Conf.Emails)

	if !yes {
		err = sendPreviewMail(mail)
		if err != nil {
			return err
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
			return fmt.Errorf("could not build confirm form: %w", err)
		}
		if !confirm {
			fmt.Printf("❌ sending aborted\n")
			os.Exit(2)
		}
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

	log.Printf("newsletter sent to %v email addresses with %v error(s)", addrCount, errCount)
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
	flag.BoolVar(&yes, "y", false, "always answer yes when program ask for confirmation")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		log.Fatalf("missing subcommand")
	}

	var cmdErr error

	switch args[0] {
	case "init":
		cmdErr = initialize()
	case "stop":
		cmdErr = stop()
	case "send":
		cmdErr = send(args[1:])
	case "preview":
		cmdErr = preview(args[1:])
	default:
		log.Fatalln("invalid sub command")
	}

	if cmdErr != nil {
		log.Fatalf("%s error: %v", args[0], cmdErr)
	}
}
