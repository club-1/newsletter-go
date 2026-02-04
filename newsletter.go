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

package newsletter

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/club-1/newsletter-go/mailer"
)

const (
	ConfigPath = ".config/newsletter"
)

type Newsletter struct {
	Config    *Config
	Hostname  string
	LocalUser string
}

// InitNewsletter initialises everything needed for the newsletter program.
// It reads the current user and its home directory than loads the config
// from the filesystem.
func InitNewsletter() (*Newsletter, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("could not get hostname: %w", err)
	}

	user, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("could not get local user: %w", err)
	}

	config, err := InitConfig(filepath.Join(user.HomeDir, ConfigPath))
	if err != nil {
		return nil, fmt.Errorf("could not init config: %w", err)
	}

	return &Newsletter{
		Config:    config,
		Hostname:  hostname,
		LocalUser: user.Username,
	}, nil
}

func (nl *Newsletter) PostmasterAddr() string {
	return "postmaster@" + nl.Hostname
}

func (nl *Newsletter) LocalUserAddr() string {
	return nl.LocalUser + "@" + nl.Hostname
}

func (nl *Newsletter) UnsubscribeAddr() string {
	return nl.LocalUser + "+" + RouteUnSubscribe + "@" + nl.Hostname
}

func (nl *Newsletter) SubscribeConfirmAddr() string {
	return nl.LocalUser + "+" + RouteSubscribeConfirm + "@" + nl.Hostname
}

func (nl *Newsletter) SendConfirmAddr() string {
	return nl.LocalUser + "+" + RouteSendConfirm + "@" + nl.Hostname
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base32.StdEncoding.EncodeToString(sum[0:32])
}

func (nl *Newsletter) HashWithSecret(s string) string {
	return hashString(s + nl.Config.Secret)
}

// generate a Message-ID
// it's based on incoming mail From address and local .secret file content
func (nl *Newsletter) GenerateId(hash string) string {
	return nl.LocalUser + "-" + hash + "@" + nl.Hostname
}

// retrive hash from message-ID using the form: `USER-HASH@SERVER`
func (nl *Newsletter) GetHashFromId(messageId string) (string, error) {
	after, prefixFound := strings.CutPrefix(messageId, nl.LocalUser+"-")
	before, suffixFound := strings.CutSuffix(after, "@"+nl.Hostname)
	if !prefixFound || !suffixFound {
		return "", fmt.Errorf("message ID does'nt match generated ID form")
	}
	return before, nil
}

func Brackets(addr string) string {
	return "<" + addr + ">"
}

// pre-fill the base mail with default values
func (nl *Newsletter) DefaultMail(subject string, body string) *mailer.Mail {
	if nl.Config.Settings.Title != "" {
		subject = "[" + nl.Config.Settings.Title + "] " + subject
	}

	if nl.Config.Signature != "" {
		body = body + "\n\n-- \n" + nl.Config.Signature
	}

	return &mailer.Mail{
		FromAddr:        nl.LocalUser + "@" + nl.Hostname,
		FromName:        nl.Config.Settings.DisplayName,
		ListUnsubscribe: fmt.Sprintf("<mailto:%s>", nl.UnsubscribeAddr()),
		Subject:         subject,
		Body:            body,
	}
}

// add a `(preview)` text after original subject
func (nl *Newsletter) SendPreviewMail(mail *mailer.Mail) error {
	mail.To = nl.LocalUserAddr()
	mail.Subject += " (preview)"

	err := mailer.Send(mail)
	if err != nil {
		return fmt.Errorf("could not send preview mail: %w", err)
	}
	fmt.Printf("📨 preview email send to %s\n", nl.LocalUserAddr())
	return nil
}

// send the newsletter to all the subscribed addresses
func (nl *Newsletter) SendNews(mail *mailer.Mail) error {
	var errCount = 0
	for _, address := range nl.Config.Emails {
		mail.To = address
		err := mailer.Send(mail)
		if err != nil {
			errCount++
		}
		time.Sleep(200 * time.Millisecond)
	}
	if errCount > 0 {
		return fmt.Errorf("error occured while sending mail to %v addresses", errCount)
	}
	return nil
}
