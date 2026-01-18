package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"slices"

	"github.com/club-1/newsletter-go"

	"github.com/mnako/letters"
)

const CmdName = "newsletterctl"

var (
	incomingEmail letters.Email
	fromAddr      string

	Messages = newsletter.Messages

	SysLog *syslog.Writer
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
		msg := fmt.Sprintf("error while sending response mail: %v", err)
		SysLog.Err(msg)
	} else {
		msg := fmt.Sprintf("response mail sent to %q", fromAddr)
		SysLog.Info(msg)
	}
}

func subscribe() error {
	if slices.Contains(newsletter.Conf.Emails, fromAddr) {
		sendResponse(
			Messages.AlreadySubscribed_subject.Print(),
			fmt.Sprintf(Messages.AlreadySubscribed_body.Print(), newsletter.PostmasterAddr()),
		)
		return fmt.Errorf("already subscribed")
	}

	var responseBody string
	if newsletter.Conf.Settings.Title == "" {
		responseBody = fmt.Sprintf(Messages.ConfirmSubscriptionAlt_body.Print(), newsletter.LocalUser)
	} else {
		responseBody = fmt.Sprintf(Messages.ConfirmSubscription_body.Print(), newsletter.Conf.Settings.Title)
	}

	mail := response(Messages.ConfirmSubscription_subject.Print(), responseBody)
	mail.ReplyTo = newsletter.SubscribeConfirmAddr()
	mail.Id = fmt.Sprintf("<%s>", newsletter.GenerateId(newsletter.HashWithSecret(fromAddr)))
	newsletter.SendMail(mail)

	msg := fmt.Sprintf("subscription confirmation mail sent to %q", fromAddr)
	SysLog.Info(msg)
	return nil
}

func subscribeConfirm() error {
	if slices.Contains(newsletter.Conf.Emails, fromAddr) {
		sendResponse(
			Messages.AlreadySubscribed_subject.Print(),
			fmt.Sprintf(Messages.AlreadySubscribed_body.Print(), newsletter.PostmasterAddr()),
		)
		return fmt.Errorf("already subscribed")
	}

	if len(incomingEmail.Headers.InReplyTo) == 0 {
		return fmt.Errorf("missing In-Reply-To header")
	}

	messageId := string(incomingEmail.Headers.InReplyTo[0])
	if messageId != newsletter.GenerateId(newsletter.HashWithSecret(fromAddr)) {
		sendResponse(
			Messages.VerificationFailed_subject.Print(),
			fmt.Sprintf(Messages.VerificationFailed_body.Print(), newsletter.LocalUserAddr()),
		)
		return fmt.Errorf("hash verification failed")
	}

	err := newsletter.Conf.Subscribe(fromAddr)
	if err != nil {
		return fmt.Errorf("error while subscribing address: %v", err)
	}
	msg := fmt.Sprintf("address %q has been added to subscribers", fromAddr)
	SysLog.Info(msg)

	var responseBody string
	if newsletter.Conf.Settings.Title == "" {
		responseBody = fmt.Sprintf(Messages.SuccessfullSubscriptionAlt_body.Print(), newsletter.LocalUser)
	} else {
		responseBody = fmt.Sprintf(Messages.SuccessfullSubscription_body.Print(), newsletter.Conf.Settings.Title)
	}

	sendResponse(Messages.SuccessfullSubscription_subject.Print(), responseBody)
	return nil
}

func unsubscribe() error {
	err := newsletter.Conf.Unsubscribe(fromAddr)
	if err != nil {
		var responseBody string
		if newsletter.Conf.Settings.Title == "" {
			responseBody = fmt.Sprintf(Messages.UnsubscriptionFailedAlt_body.Print(), newsletter.LocalUser, newsletter.LocalUserAddr())
		} else {
			responseBody = fmt.Sprintf(Messages.UnsubscriptionFailed_body.Print(), newsletter.Conf.Settings.Title, newsletter.LocalUserAddr())
		}
		sendResponse(Messages.UnsubscriptionFailed_subject.Print(), responseBody)
		return fmt.Errorf("could not unsubscribe: %w", err)
	}

	msg := fmt.Sprintf("address %q removed from subscribers", fromAddr)
	SysLog.Info(msg)

	var responseBody string
	if newsletter.Conf.Settings.Title == "" {
		responseBody = fmt.Sprintf(Messages.SuccessfullUnsubscriptionAlt_body.Print(), newsletter.LocalUser)
	} else {
		responseBody = fmt.Sprintf(Messages.SuccessfullUnsubscription_body.Print(), newsletter.Conf.Settings.Title)
	}

	sendResponse(Messages.SuccessfullUnsubscription_subject.Print(), responseBody)
	return nil
}

func send() error {
	if fromAddr != newsletter.LocalUserAddr() {
		return fmt.Errorf("email From does'nt match user address")
	}

	body := incomingEmail.Text
	subject := incomingEmail.Headers.Subject

	hash := newsletter.HashWithSecret(body + subject)

	bodyFilePath := filepath.Join(os.TempDir(), "newsletter-send-"+hash+".body.txt")
	subjectFilePath := filepath.Join(os.TempDir(), "newsletter-send-"+hash+".subject.txt")

	var err error
	err = os.WriteFile(bodyFilePath, []byte(body), 0660)
	if err != nil {
		return err
	}
	err = os.WriteFile(subjectFilePath, []byte(subject), 0660)
	if err != nil {
		return err
	}

	mail := newsletter.DefaultMail(subject, body)
	mail.Id = fmt.Sprintf("<%s>", newsletter.GenerateId(hash))
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", newsletter.UnsubscribeAddr())
	mail.Body += fmt.Sprintf("(\n\nthis is a preview mail, if you want to confirm and send the newsletter to all the %v subscribers, reply to this email)", len(newsletter.Conf.Emails))
	mail.ReplyTo = newsletter.SendConfirmAddr()

	return newsletter.SendPreviewMail(mail)
}

func sendConfirm() error {
	if fromAddr != newsletter.LocalUserAddr() {
		return fmt.Errorf("email From header does'nt match user address")
	}

	if len(incomingEmail.Headers.InReplyTo) == 0 {
		return fmt.Errorf("missing In-Reply-To header")
	}

	messageId := string(incomingEmail.Headers.InReplyTo[0])
	hash, err := newsletter.GetHashFromId(messageId)
	if err != nil {
		return fmt.Errorf("In-Reply-To parsing error: %w", err)
	}

	bodyFilePath := filepath.Join(os.TempDir(), "newsletter-send-"+hash+".body.txt")
	subjectFilePath := filepath.Join(os.TempDir(), "newsletter-send-"+hash+".subject.txt")

	var body string
	_, err = os.Stat(bodyFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("hash does not match existing temporary body file")
	} else {
		bodyB, err := os.ReadFile(bodyFilePath)
		if err != nil {
			return fmt.Errorf("could read temporary body file: %w", err)
		}
		body = string(bodyB)
	}

	var subject string
	_, err = os.Stat(subjectFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("hash does not match existing temporary subject file")
	} else {
		subjectB, err := os.ReadFile(subjectFilePath)
		if err != nil {
			return fmt.Errorf("could read temporary subject file: %w", err)
		}
		subject = string(subjectB)
	}

	mail := newsletter.DefaultMail(subject, body)
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", newsletter.UnsubscribeAddr())
	err = newsletter.SendNews(mail)
	if err != nil {
		return fmt.Errorf("sending newsletter: %w", err)
	}
	msg := fmt.Sprintf("newsletter successfully send to all the %v subscribers", len(newsletter.Conf.Emails))
	SysLog.Info(msg)
	return nil
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("missing sub command")
	}

	var err error
	SysLog, err = syslog.New(syslog.LOG_USER, CmdName)
	if err != nil {
		log.Fatal(err)
	}

	// IMPORTANT: cannot use `:=` beccause it need to setup global var `incomingEmail`
	incomingEmail, err = letters.ParseEmail(os.Stdin)
	if err != nil {
		msg := fmt.Sprintf("error while parsing input email: %v", err)
		SysLog.Err(msg)
		os.Exit(1)
	}

	err = newsletter.ReadConfig()
	if err != nil {
		msg := fmt.Sprintf("init: %v", err)
		SysLog.Err(msg)
		os.Exit(1)
	}

	cmdErrPrefix := fmt.Sprintf("recieved mail to route %q", args[0])

	if len(incomingEmail.Headers.From) == 0 {
		msg := fmt.Sprintf(cmdErrPrefix + " without From header")
		SysLog.Err(msg)
		os.Exit(1)
	}

	fromAddr = incomingEmail.Headers.From[0].Address
	cmdErrPrefix += fmt.Sprintf(" from %q", fromAddr)

	var cmdErr error

	switch args[0] {
	case newsletter.RouteSubscribe:
		cmdErr = subscribe()
	case newsletter.RouteSubscribeConfirm:
		cmdErr = subscribeConfirm()
	case newsletter.RouteUnSubscribe:
		cmdErr = unsubscribe()
	case newsletter.RouteSend:
		cmdErr = send()
	case newsletter.RouteSendConfirm:
		cmdErr = sendConfirm()
	default:
		msg := fmt.Sprintf("invalid sub command: %q", args[0])
		SysLog.Err(msg)
		os.Exit(1)
	}

	if cmdErr != nil {
		msg := fmt.Sprintf(cmdErrPrefix+", error: %v", cmdErr)
		SysLog.Err(msg)
		// do not send non-zero response code because otherwise
		// it would answer an error feedback automatically by email
	}
}
