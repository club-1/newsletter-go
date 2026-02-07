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
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"slices"

	"github.com/club-1/newsletter-go/v3"
	"github.com/club-1/newsletter-go/v3/mailer"
	"github.com/club-1/newsletter-go/v3/messages"

	"github.com/mnako/letters"
)

const CmdName = "newsletterctl"

var (
	sysLog        *syslog.Writer
	nl            *newsletter.Newsletter
	incomingEmail letters.Email
	fromAddr      string
)

func sysLogErr(msg string) {
	sysLog.Err(fmt.Sprintf("%v: %v", nl.LocalUser, msg))
}

func sysLogInfo(msg string) {
	sysLog.Info(fmt.Sprintf("%v: %v", nl.LocalUser, msg))
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
			messages.AlreadySubscribed_subject.Print(),
			fmt.Sprintf(messages.AlreadySubscribed_body.Print(), nl.PostmasterAddr()),
		)
		return fmt.Errorf("already subscribed")
	}

	var responseBody string
	if nl.Config.Settings.Title == "" {
		responseBody = fmt.Sprintf(messages.ConfirmSubscriptionAlt_body.Print(), nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(messages.ConfirmSubscription_body.Print(), nl.Config.Settings.Title)
	}

	mail := response(messages.ConfirmSubscription_subject.Print(), responseBody)
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
			messages.AlreadySubscribed_subject.Print(),
			fmt.Sprintf(messages.AlreadySubscribed_body.Print(), nl.PostmasterAddr()),
		)
		return fmt.Errorf("already subscribed")
	}

	if len(incomingEmail.Headers.InReplyTo) == 0 {
		return fmt.Errorf("missing In-Reply-To header")
	}

	messageId := string(incomingEmail.Headers.InReplyTo[0])
	if messageId != nl.GenerateId(nl.HashWithSecret(fromAddr)) {
		sendResponse(
			messages.VerificationFailed_subject.Print(),
			fmt.Sprintf(messages.VerificationFailed_body.Print(), nl.LocalUserAddr()),
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
		responseBody = fmt.Sprintf(messages.SuccessfullSubscriptionAlt_body.Print(), nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(messages.SuccessfullSubscription_body.Print(), nl.Config.Settings.Title)
	}

	sendResponse(messages.SuccessfullSubscription_subject.Print(), responseBody)
	return nil
}

func unsubscribe() error {
	err := nl.Config.Unsubscribe(fromAddr)
	if err != nil {
		var responseBody string
		if nl.Config.Settings.Title == "" {
			responseBody = fmt.Sprintf(messages.UnsubscriptionFailedAlt_body.Print(), nl.LocalUser, nl.LocalUserAddr())
		} else {
			responseBody = fmt.Sprintf(messages.UnsubscriptionFailed_body.Print(), nl.Config.Settings.Title, nl.LocalUserAddr())
		}
		sendResponse(messages.UnsubscriptionFailed_subject.Print(), responseBody)
		return fmt.Errorf("could not unsubscribe: %w", err)
	}

	msg := fmt.Sprintf("address %q removed from subscribers", fromAddr)
	sysLogInfo(msg)

	var responseBody string
	if nl.Config.Settings.Title == "" {
		responseBody = fmt.Sprintf(messages.SuccessfullUnsubscriptionAlt_body.Print(), nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(messages.SuccessfullUnsubscription_body.Print(), nl.Config.Settings.Title)
	}

	sendResponse(messages.SuccessfullUnsubscription_subject.Print(), responseBody)
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
			return fmt.Errorf("read temporary body file: %w", err)
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
			return fmt.Errorf("read temporary subject file: %w", err)
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
	sysLog, err = syslog.New(syslog.LOG_USER, newsletter.LogIdentifier)
	if err != nil {
		log.Fatal(err)
	}

	nl, err = newsletter.InitNewsletter()
	if err != nil {
		msg := fmt.Sprintf("init: %v", err)
		sysLog.Crit(msg)
		os.Exit(1)
	}
	messages.SetLanguage(nl.Config.Settings.Language)

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
