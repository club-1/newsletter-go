package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/club-1/newsletter-go"
)

const Name = "newsletter"

var verbose bool
var l = log.New(os.Stderr, Name+": ", 0)

func getPrefix() (string, error) {
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

func PrintPreview(mail *newsletter.Mail) {
	fmt.Print("================ PREVIEW START ================\n")
	fmt.Print("┌---- Header ------\n")
	fmt.Printf("| Subject: %s\n", mail.Subject)
	fmt.Printf("| From: %s\n", mail.From())
	fmt.Print("└------------------\n")
	fmt.Printf("%s\n", mail.Body)
	fmt.Print("================  PREVIEW END  ================\n")
}

func Init() {
	prefix, err := getPrefix()
	if err != nil {
		l.Fatalln("cannot get prefix:", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		l.Fatalln("cannot get user home directory:", err)
	}
	routes := [5]string{"subscribe", "unsubscribe", "subscribe-confirm", "send", "send-confirm"}
	for _, route := range routes {
		fileName := ".forward+" + route
		filePath := filepath.Join(homeDir, fileName)
		if verbose {
			fmt.Println("writting file", filePath)
		}

		cmdPath := filepath.Join(prefix, "sbin/newsletterctl")
		content := []byte("| \"" + cmdPath + " " + route + "\"\n")
		err := os.WriteFile(filePath, content, 0664)
		if err != nil {
			l.Println("cannot write file:", filePath)
		}
	}
}

func Send(args []string) {
	if len(args) < 2 {
		log.Fatalf("missing arguments")
	}
	if len(args) > 2 {
		log.Fatalf("too many arguments")
	}
	subject := args[0]
	bodyPath := args[1]
	bodyB, err := os.ReadFile(bodyPath)
	if err != nil {
		log.Fatalf("could not load newsletter body: %w", err)
	}
	mail := newsletter.DefaultMail(subject, string(bodyB))
	PrintPreview(mail)

	fmt.Printf("\nDo you really want to send this to %v email addresses ?\n", len(newsletter.Conf.Emails))
}

func main() {
	log.SetFlags(0)
	flag.BoolVar(&verbose, "v", false, "increase verbosity of program")

	user, err := user.Current()
	if err != nil {
		log.Fatalf("could not get local user: %v", err)
	}
	newsletter.LocalUser = user.Username

	err = newsletter.ReadConfig()
	if err != nil {
		log.Printf("init: %v", err)
	}

	flag.Parse()

	args := flag.Args()

	if len(args) >= 1 {
		switch args[0] {
		case "init":
			Init()
		case "send":
			Send(args[1:])
		default:
			l.Fatalln("invalid sub command")
		}
	}
}
