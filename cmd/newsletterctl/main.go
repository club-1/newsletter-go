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
	"github.com/club-1/newsletter-go/mailer"

	"github.com/mnako/letters"
)

const CmdName = "newsletterctl"

var (
	SysLog        *syslog.Writer
	nl            *newsletter.Newsletter
	incomingEmail letters.Email
	fromAddr      string

	Messages = newsletter.Messages
)

func sysLogErr(msg string) {
	SysLog.Err(fmt.Sprintf("%v: %v", nl.LocalUser, msg))
}

func sysLogInfo(msg string) {
	SysLog.Info(fmt.Sprintf("%v: %v", nl.LocalUser, msg))
}

// base response mail directed toward recieved From address
func response(subject string, body string) *mailer.Mail {
	mail := nl.DefaultMail(subject, body)

	messageId := string(incomingEmail.Headers.MessageID)
	mail.InReplyTo = newsletter.Brackets(messageId)
	mail.To = fromAddr

	return mail
}

// Send standard response and log
func sendResponse(subject string, body string) {
	mail := response(subject, body)
	err := mailer.Send(mail)
	if err != nil {
		msg := fmt.Sprintf("error while sending response mail: %v", err)
		sysLogErr(msg)
	} else {
		msg := fmt.Sprintf("response mail sent to %q", fromAddr)
		sysLogInfo(msg)
	}
}

func subscribe() error {
	if slices.Contains(nl.Config.Emails, fromAddr) {
		sendResponse(
			Messages.AlreadySubscribed_subject.Print(),
			fmt.Sprintf(Messages.AlreadySubscribed_body.Print(), nl.PostmasterAddr()),
		)
		return fmt.Errorf("already subscribed")
	}

	var responseBody string
	if nl.Config.Settings.Title == "" {
		responseBody = fmt.Sprintf(Messages.ConfirmSubscriptionAlt_body.Print(), nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(Messages.ConfirmSubscription_body.Print(), nl.Config.Settings.Title)
	}

	mail := response(Messages.ConfirmSubscription_subject.Print(), responseBody)
	mail.ReplyTo = nl.SubscribeConfirmAddr()
	mail.Id = fmt.Sprintf("<%s>", nl.GenerateId(nl.HashWithSecret(fromAddr)))
	mailer.Send(mail)

	msg := fmt.Sprintf("subscription confirmation mail sent to %q", fromAddr)
	sysLogInfo(msg)
	return nil
}

func subscribeConfirm() error {
	if slices.Contains(nl.Config.Emails, fromAddr) {
		sendResponse(
			Messages.AlreadySubscribed_subject.Print(),
			fmt.Sprintf(Messages.AlreadySubscribed_body.Print(), nl.PostmasterAddr()),
		)
		return fmt.Errorf("already subscribed")
	}

	if len(incomingEmail.Headers.InReplyTo) == 0 {
		return fmt.Errorf("missing In-Reply-To header")
	}

	messageId := string(incomingEmail.Headers.InReplyTo[0])
	if messageId != nl.GenerateId(nl.HashWithSecret(fromAddr)) {
		sendResponse(
			Messages.VerificationFailed_subject.Print(),
			fmt.Sprintf(Messages.VerificationFailed_body.Print(), nl.LocalUserAddr()),
		)
		return fmt.Errorf("hash verification failed")
	}

	err := nl.Config.Subscribe(fromAddr)
	if err != nil {
		return fmt.Errorf("error while subscribing address: %v", err)
	}
	msg := fmt.Sprintf("address %q has been added to subscribers", fromAddr)
	sysLogInfo(msg)

	var responseBody string
	if nl.Config.Settings.Title == "" {
		responseBody = fmt.Sprintf(Messages.SuccessfullSubscriptionAlt_body.Print(), nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(Messages.SuccessfullSubscription_body.Print(), nl.Config.Settings.Title)
	}

	sendResponse(Messages.SuccessfullSubscription_subject.Print(), responseBody)
	return nil
}

func unsubscribe() error {
	err := nl.Config.Unsubscribe(fromAddr)
	if err != nil {
		var responseBody string
		if nl.Config.Settings.Title == "" {
			responseBody = fmt.Sprintf(Messages.UnsubscriptionFailedAlt_body.Print(), nl.LocalUser, nl.LocalUserAddr())
		} else {
			responseBody = fmt.Sprintf(Messages.UnsubscriptionFailed_body.Print(), nl.Config.Settings.Title, nl.LocalUserAddr())
		}
		sendResponse(Messages.UnsubscriptionFailed_subject.Print(), responseBody)
		return fmt.Errorf("could not unsubscribe: %w", err)
	}

	msg := fmt.Sprintf("address %q removed from subscribers", fromAddr)
	sysLogInfo(msg)

	var responseBody string
	if nl.Config.Settings.Title == "" {
		responseBody = fmt.Sprintf(Messages.SuccessfullUnsubscriptionAlt_body.Print(), nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(Messages.SuccessfullUnsubscription_body.Print(), nl.Config.Settings.Title)
	}

	sendResponse(Messages.SuccessfullUnsubscription_subject.Print(), responseBody)
	return nil
}

func send() error {
	if fromAddr != nl.LocalUserAddr() {
		return fmt.Errorf("email From does'nt match user address")
	}

	body := incomingEmail.Text
	subject := incomingEmail.Headers.Subject

	hash := nl.HashWithSecret(body + subject)

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

	mail := nl.DefaultMail(subject, body)
	mail.Id = fmt.Sprintf("<%s>", nl.GenerateId(hash))
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", nl.UnsubscribeAddr())
	mail.Body += fmt.Sprintf("(\n\nthis is a preview mail, if you want to confirm and send the newsletter to all the %v subscribers, reply to this email)", len(nl.Config.Emails))
	mail.ReplyTo = nl.SendConfirmAddr()

	return nl.SendPreviewMail(mail)
}

func sendConfirm() error {
	if fromAddr != nl.LocalUserAddr() {
		return fmt.Errorf("email From header does'nt match user address")
	}

	if len(incomingEmail.Headers.InReplyTo) == 0 {
		return fmt.Errorf("missing In-Reply-To header")
	}

	messageId := string(incomingEmail.Headers.InReplyTo[0])
	hash, err := nl.GetHashFromId(messageId)
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

	mail := nl.DefaultMail(subject, body)
	mail.Body += fmt.Sprintf("\n\nTo unsubscribe, send a mail to <%s>", nl.UnsubscribeAddr())
	err = nl.SendNews(mail)
	if err != nil {
		return fmt.Errorf("sending newsletter: %w", err)
	}
	msg := fmt.Sprintf("newsletter successfully send to all the %v subscribers", len(nl.Config.Emails))
	sysLogInfo(msg)
	return nil
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("missing sub command")
	}

	var err error
	SysLog, err = syslog.New(syslog.LOG_USER, newsletter.LogIdentifier)
	if err != nil {
		log.Fatal(err)
	}

	nl, err = newsletter.InitNewsletter()
	if err != nil {
		msg := fmt.Sprintf("init: %v", err)
		SysLog.Crit(msg)
		os.Exit(1)
	}

	incomingEmail, err = letters.ParseEmail(os.Stdin)
	if err != nil {
		msg := fmt.Sprintf("error while parsing input email: %v", err)
		sysLogErr(msg)
		os.Exit(1)
	}

	cmdErrPrefix := fmt.Sprintf("recieved mail to route %q", args[0])

	if len(incomingEmail.Headers.From) == 0 {
		msg := fmt.Sprintf(cmdErrPrefix + " without From header")
		sysLogErr(msg)
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
		sysLogErr(msg)
		os.Exit(1)
	}

	if cmdErr != nil {
		msg := fmt.Sprintf(cmdErrPrefix+", error: %v", cmdErr)
		sysLogErr(msg)
		// do not send non-zero response code because otherwise
		// it would answer an error feedback automatically by email
	}
}
