package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/club-1/newsletter-go"

	"github.com/mnako/letters"
)

const Name = "newsletterctl"

const (
	RouteSubscribe        = "subscribe"
	RouteSubscribeConfirm = "subscribe-confirm"
	RouteUnSubscribe      = "unsubscribe"
	RouteSend             = "send"
	RouteSendConfirm      = "send-confirm"
)

var (
	incomingEmail letters.Email
	logMessage    string
	fromAddr      string
)

func response(subject string, body string) *newsletter.Mail {
	return &newsletter.Mail{
		To:        fromAddr,
		FromAddr:  newsletter.LocalUser + "@" + newsletter.LocalServer,
		FromName:  newsletter.Conf.Settings.DisplayName,
		Subject:   "[" + newsletter.Conf.Settings.DisplayName + "] " + subject,
		Body:      body + "--\n" + newsletter.Conf.Signature,
		InReplyTo: string(incomingEmail.Headers.MessageID),
	}
}

func Subscribe() {
	if slices.Contains(newsletter.Conf.Emails, fromAddr) {
		log.Printf(logMessage + ": already subscribed")
		postmaster := newsletter.Brackets(newsletter.PostmasterAddr())
		mail := response("already subscribed", "your email is already subscribed, if problem persist, contact "+postmaster)
		newsletter.SendMail(mail)
		log.Printf("subscription state info mail sent to %q", fromAddr)
	} else {
		log.Printf(logMessage + ": unsubscribed")
		log.Printf("subscription confirmation mail sent to %q", fromAddr)
	}
}

func SubscribeConfirm() {
	log.Println("recieved mail to route 'subscribe-confirm'")
}

func Unsubscribe() {
	err := newsletter.Conf.Unsubscribe(fromAddr)
	if err != nil {
		log.Printf(logMessage+": %v", err)
	} else {
		log.Printf(logMessage + ": successfully unsubscribed")
	}
}

func Send() {
	log.Println("recieved mail to route 'send'")
}

func SendConfirm() {
	log.Println("recieved mail to route 'send-confirm'")
}

func initLogger() *os.File {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalln("cannot get user cache directory:", err)
	}

	logDir := filepath.Join(userCacheDir, "newsletter")

	err = os.MkdirAll(logDir, 0775)
	if err != nil {
		log.Fatalln("cannot create log folder:", err)
	}
	LogFilePath := filepath.Join(logDir, Name+".log")

	logFile, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	if err != nil {
		panic(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	return logFile
}

func main() {
	logFile := initLogger()
	defer logFile.Close()

	incomingEmail, err := letters.ParseEmail(os.Stdin)
	if err != nil {
		log.Fatalf("error while parsing input email: %v", err)
	}

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("missing sub command")
	}

	logMessage += fmt.Sprintf("recieved mail to route %q", args[0])

	if len(incomingEmail.Headers.From) == 0 {
		log.Fatalf(logMessage + " without From header")
	}

	err = newsletter.ReadConfig()
	if err != nil {
		log.Fatalf("init: %v", err)
	}

	fromAddr = incomingEmail.Headers.From[0].Address
	logMessage += fmt.Sprintf(" from %q", fromAddr)

	switch args[0] {
	case RouteSubscribe:
		Subscribe()
	case RouteSubscribeConfirm:
		SubscribeConfirm()
	case RouteUnSubscribe:
		Unsubscribe()
	case RouteSend:
		Send()
	case RouteSendConfirm:
		SendConfirm()
	default:
		log.Fatalf("invalid sub command: %q", args[0])
	}
}
