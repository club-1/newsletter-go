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

var (
	incomingEmail letters.Email
	logMessage    string
	fromAddr      string
)

// base response mail directed toward recieved From address
func response(subject string, body string) *newsletter.Mail {
	mail := newsletter.DefaultMail(subject, body)

	messageId := string(incomingEmail.Headers.MessageID)
	mail.InReplyTo = newsletter.Brackets(messageId)
	mail.To = fromAddr

	return mail
}

// Send standard response and log
func sendResponse(subject string, body string) {
	mail := response(subject, body)
	err := newsletter.SendMail(mail)
	if err != nil {
		log.Printf("error while sending response mail: %v", err)
	} else {
		log.Printf("response mail sent to %q", fromAddr)
	}
}

func Subscribe() {
	if slices.Contains(newsletter.Conf.Emails, fromAddr) {
		log.Printf(logMessage + ": already subscribed")
		postmaster := fmt.Sprintf("<%s>", newsletter.PostmasterAddr())
		sendResponse("already subscribed", "your email is already subscribed, if problem persist, contact "+postmaster)
	} else {
		log.Printf(logMessage + ": unsubscribed")
		mail := response("confirm your subsciption", "Reply to this email to confirm that you want to subscribe to "+newsletter.Conf.Settings.Title)
		mail.ReplyTo = newsletter.SubscribeConfirmAddr()
		mail.Id = fmt.Sprintf("<%s>", newsletter.GenerateId(fromAddr))
		newsletter.SendMail(mail)
		log.Printf("subscription confirmation mail sent to %q", fromAddr)
	}
}

func SubscribeConfirm() error {
	if slices.Contains(newsletter.Conf.Emails, fromAddr) {
		sendResponse(
			"already subscribed",
			fmt.Sprintf("your email is already subscribed, if problem persist, contact <%s>", newsletter.PostmasterAddr()),
		)
		return fmt.Errorf("already subscribed")
	}

	if len(incomingEmail.Headers.InReplyTo) == 0 {
		return fmt.Errorf("missing In-Replay-To header")
	}

	messageId := string(incomingEmail.Headers.InReplyTo[0])
	if messageId != newsletter.GenerateId(fromAddr) {
		sendResponse("an error occured", "your email cannot be added to the subscripted list, contact list owner for more infos")
		return fmt.Errorf("hash verification failed")
	}

	err := newsletter.Conf.Subscribe(fromAddr)
	if err != nil {
		return fmt.Errorf("error while subscribing address: %v", err)
	}
	log.Printf("address %q has been added to subscribers", fromAddr)
	sendResponse("subscription is successfull", "your email has been added to list "+newsletter.Conf.Settings.Title)
	return nil
}

func Unsubscribe() {
	err := newsletter.Conf.Unsubscribe(fromAddr)
	if err != nil {
		log.Printf(logMessage+": %v", err)
	} else {
		log.Printf(logMessage + ": successfully unsubscribed")
		sendResponse("successfully unsubscribed", "your email was successfully removed from the list "+newsletter.Conf.Settings.Title)
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

	var err error // IMPORTANT: cannot use `:=` beccause it need to setup global var `incomingEmail`
	incomingEmail, err = letters.ParseEmail(os.Stdin)
	if err != nil {
		log.Fatalf("error while parsing input email: %v", err)
	}

	err = newsletter.ReadConfig()
	if err != nil {
		log.Fatalf("init: %v", err)
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

	fromAddr = incomingEmail.Headers.From[0].Address
	logMessage += fmt.Sprintf(" from %q", fromAddr)

	var cmdErr error

	switch args[0] {
	case newsletter.RouteSubscribe:
		Subscribe()
	case newsletter.RouteSubscribeConfirm:
		cmdErr = SubscribeConfirm()
	case newsletter.RouteUnSubscribe:
		Unsubscribe()
	case newsletter.RouteSend:
		Send()
	case newsletter.RouteSendConfirm:
		SendConfirm()
	default:
		log.Fatalf("invalid sub command: %q", args[0])
	}

	if cmdErr != nil {
		log.Fatalf(logMessage+", error: %v", cmdErr)
	}
}
