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

package control

import (
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"log/syslog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/club-1/newsletter-go/v3"
	"github.com/club-1/newsletter-go/v3/mailer"
	"github.com/club-1/newsletter-go/v3/messages"
)

const (
	logIdentifier = "newsletter"
)

type Controller struct {
	logger *slog.Logger
	nl     *newsletter.Newsletter
}

func NewController() (*Controller, error) {
	sysLog, err := syslog.New(syslog.LOG_USER, logIdentifier)
	if err != nil {
		return nil, fmt.Errorf("init syslog: %w", err)
	}
	logger := slog.New(NewSyslogHandler(sysLog))

	nl, err := newsletter.New()
	if err != nil {
		logger.Error(fmt.Sprintf("init newsletter: %v", err))
		return nil, err
	}
	messages.SetLanguage(nl.Config.Settings.Language)

	logger = logger.With("user", nl.LocalUser)

	return &Controller{
		logger: logger,
		nl:     nl,
	}, nil
}

// response creates a new [mailer.Mail] directed towards the request's From
// address.
func (c *Controller) response(req *Request, subject string, body string) *mailer.Mail {
	mail := c.nl.DefaultMail(subject, body)
	mail.InReplyTo = fmt.Sprintf("<%s>", req.MessageID)
	mail.To = req.From.Address

	referencesBuilder := strings.Builder{}
	for _, id := range req.Headers.References {
		fmt.Fprintf(&referencesBuilder, "<%s> ", id)
	}
	fmt.Fprintf(&referencesBuilder, "<%s>", req.MessageID)
	mail.References = referencesBuilder.String()

	return mail
}

// sendResponse sends a reply to the received mail and logs the result.
func (c *Controller) sendResponse(req *Request, subject string, body string) {
	mail := c.response(req, subject, body)
	err := c.nl.Mailer.Send(mail)
	if err != nil {
		req.Log.Error(fmt.Sprintf("error while sending response mail: %v", err))
	} else {
		req.Log.Info("response mail sent")
	}
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base32.StdEncoding.EncodeToString(sum[0:32])
}

func (c *Controller) HashWithSecret(s string) string {
	return hashString(s + c.nl.Config.Secret)
}

// GenerateId generates a Message-ID for this newsletter using the given hash.
func (c *Controller) GenerateId(hash string) string {
	return fmt.Sprintf("%s-%s@%s", c.nl.LocalUser, hash, c.nl.Hostname)
}

func (c *Controller) GenerateConfirmID(req *Request) string {
	hash := c.HashWithSecret(req.From.Address)
	return c.GenerateId(hash)
}

// GetHashFromId retrieves the hash from the given messageID of the form: `USER-HASH@SERVER`
func (c *Controller) GetHashFromId(messageID string) (string, error) {
	after, prefixFound := strings.CutPrefix(messageID, c.nl.LocalUser+"-")
	before, suffixFound := strings.CutSuffix(after, "@"+c.nl.Hostname)
	if !prefixFound || !suffixFound {
		return "", errors.New("message ID doesn't match generated ID form")
	}
	return before, nil
}

func (c *Controller) subscribe(req *Request) error {
	if slices.Contains(c.nl.Config.Emails, req.From.Address) {
		req.Log.Warn("address is already subscribed")
		c.sendResponse(
			req,
			messages.AlreadySubscribed_subject.Print(),
			fmt.Sprintf(messages.AlreadySubscribed_body.Print(), c.nl.PostmasterAddr()),
		)
		return nil
	}

	var responseBody string
	if c.nl.Config.Settings.Title == "" {
		responseBody = fmt.Sprintf(messages.ConfirmSubscriptionAlt_body.Print(), c.nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(messages.ConfirmSubscription_body.Print(), c.nl.Config.Settings.Title)
	}

	mail := c.response(req, messages.ConfirmSubscription_subject.Print(), responseBody)
	mail.ReplyTo = c.nl.SubscribeConfirmAddr()
	mail.Id = fmt.Sprintf("<%s>", c.GenerateConfirmID(req))

	err := c.nl.Mailer.Send(mail)
	if err != nil {
		return fmt.Errorf("send response mail: %v", err)
	}

	req.Log.Info("subscription confirmation mail sent")
	return nil
}

func (c *Controller) subscribeConfirm(req *Request) error {
	if slices.Contains(c.nl.Config.Emails, req.From.Address) {
		req.Log.Warn("address is already subscribed")
		c.sendResponse(
			req,
			messages.AlreadySubscribed_subject.Print(),
			fmt.Sprintf(messages.AlreadySubscribed_body.Print(), c.nl.PostmasterAddr()),
		)
		return nil
	}

	if len(req.Headers.InReplyTo) == 0 {
		return fmt.Errorf("missing In-Reply-To header")
	}

	messageId := string(req.Headers.InReplyTo[0])
	if messageId != c.GenerateConfirmID(req) {
		c.sendResponse(
			req,
			messages.VerificationFailed_subject.Print(),
			fmt.Sprintf(messages.VerificationFailed_body.Print(), c.nl.LocalUserAddr()),
		)
		return fmt.Errorf("hash verification failed")
	}

	err := c.nl.Config.Subscribe(req.From.Address)
	if err != nil {
		return fmt.Errorf("error while subscribing address: %v", err)
	}
	req.Log.Info("address has been added to subscribers")

	var responseBody string
	if c.nl.Config.Settings.Title == "" {
		responseBody = fmt.Sprintf(messages.SuccessfullSubscriptionAlt_body.Print(), c.nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(messages.SuccessfullSubscription_body.Print(), c.nl.Config.Settings.Title)
	}

	c.sendResponse(req, messages.SuccessfullSubscription_subject.Print(), responseBody)
	return nil
}

func (c *Controller) unsubscribe(req *Request) error {
	err := c.nl.Config.Unsubscribe(req.From.Address)
	switch {
	case err == nil:
		req.Log.Info("address removed from subscribers")
	case errors.Is(err, newsletter.ErrNotSubscribed):
		req.Log.Warn("address is not subscribed")
	default:
		var responseBody string
		if c.nl.Config.Settings.Title == "" {
			responseBody = fmt.Sprintf(messages.UnsubscriptionFailedAlt_body.Print(), c.nl.LocalUser, c.nl.LocalUserAddr())
		} else {
			responseBody = fmt.Sprintf(messages.UnsubscriptionFailed_body.Print(), c.nl.Config.Settings.Title, c.nl.LocalUserAddr())
		}
		c.sendResponse(req, messages.UnsubscriptionFailed_subject.Print(), responseBody)
		return fmt.Errorf("could not unsubscribe: %w", err)
	}

	var responseBody string
	if c.nl.Config.Settings.Title == "" {
		responseBody = fmt.Sprintf(messages.SuccessfullUnsubscriptionAlt_body.Print(), c.nl.LocalUser)
	} else {
		responseBody = fmt.Sprintf(messages.SuccessfullUnsubscription_body.Print(), c.nl.Config.Settings.Title)
	}

	c.sendResponse(req, messages.SuccessfullUnsubscription_subject.Print(), responseBody)
	return nil
}

func (c *Controller) send(req *Request) error {
	if req.From.Address != c.nl.LocalUserAddr() {
		return fmt.Errorf("email From doesn't match user address")
	}

	body := req.Text
	subject := req.Headers.Subject

	hash := c.HashWithSecret(body + subject)

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

	mail := c.nl.DefaultMail(subject, body)
	mail.Id = c.GenerateId(hash)
	mail.Body += fmt.Sprintf(messages.Newsletter_footer.Print(), c.nl.UnsubscribeAddr())
	mail.Body += fmt.Sprintf("\n\n(this is a preview mail, if you want to confirm and send the newsletter to all the %v subscribers, reply to this email)", len(c.nl.Config.Emails))
	mail.ReplyTo = c.nl.SendConfirmAddr()

	return c.nl.SendPreviewMail(*mail)
}

func (c *Controller) sendConfirm(req *Request) error {
	if req.From.Address != c.nl.LocalUserAddr() {
		return fmt.Errorf("email From header doesn't match user address")
	}

	if len(req.Headers.InReplyTo) == 0 {
		return fmt.Errorf("missing In-Reply-To header")
	}

	messageId := string(req.Headers.InReplyTo[0])
	hash, err := c.GetHashFromId(messageId)
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

	mail := c.nl.DefaultMail(subject, body)
	mail.Body += fmt.Sprintf(messages.Newsletter_footer.Print(), c.nl.UnsubscribeAddr())
	err = c.nl.SendNews(mail)
	if err != nil {
		return fmt.Errorf("sending newsletter: %w", err)
	}
	req.Log.Info("newsletter successfully sent to all subscribers", "sent", len(c.nl.Config.Emails))
	return nil
}

func (c *Controller) Handle(route string, r io.Reader) error {
	logger := c.logger.With("route", route)

	request, err := ParseRequest(logger, r)
	if err != nil {
		logger.Error(fmt.Sprintf("parse email: %v", err))
		return err // TODO: maybe here return a better error
	}

	var cmdErr error

	switch route {
	case newsletter.RouteSubscribe:
		cmdErr = c.subscribe(request)
	case newsletter.RouteSubscribeConfirm:
		cmdErr = c.subscribeConfirm(request)
	case newsletter.RouteUnSubscribe:
		cmdErr = c.unsubscribe(request)
	case newsletter.RouteSend:
		cmdErr = c.send(request)
	case newsletter.RouteSendConfirm:
		cmdErr = c.sendConfirm(request)
	default:
		request.Log.Error("invalid route")
	}

	if cmdErr != nil {
		request.Log.Error(fmt.Sprintf("error: %v", cmdErr))
	}

	return cmdErr
}
