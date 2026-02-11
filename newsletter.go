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
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/club-1/newsletter-go/v3/mailer"
)

const (
	ConfigPath = ".config/newsletter"

	RouteSubscribe        = "subscribe"
	RouteSubscribeConfirm = "subscribe-confirm"
	RouteUnSubscribe      = "unsubscribe"
	RouteSend             = "send"
	RouteSendConfirm      = "send-confirm"
)

var (
	Routes = [...]string{RouteSubscribe, RouteSubscribeConfirm, RouteUnSubscribe, RouteSend, RouteSendConfirm}
)

type Newsletter struct {
	Config    *Config
	Hostname  string
	LocalUser string
	Mailer    mailer.Mailer
}

// New creates a new [Newsletter] instance and initialises it.
//
// It reads information about the system, the current user and its config
// directory, then loads the config from the filesystem.
func New() (*Newsletter, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("get hostname: %w", err)
	}

	user, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("get local user: %w", err)
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = user.HomeDir
	}

	config, err := InitConfig(filepath.Join(homeDir, ConfigPath))
	if err != nil {
		return nil, fmt.Errorf("init config: %w", err)
	}

	return &Newsletter{
		Config:    config,
		Hostname:  hostname,
		LocalUser: user.Username,
		Mailer:    mailer.Default(),
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

	err := nl.Mailer.Send(mail)
	if err != nil {
		return fmt.Errorf("send preview mail: %w", err)
	}
	fmt.Printf("📨 preview email send to %s\n", nl.LocalUserAddr())
	return nil
}

// send the newsletter to all the subscribed addresses
func (nl *Newsletter) SendNews(mail *mailer.Mail) error {
	var errCount = 0
	for _, address := range nl.Config.Emails {
		mail.To = address
		err := nl.Mailer.Send(mail)
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
