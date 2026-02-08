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
	"reflect"
	"strings"
	"testing"

	"github.com/club-1/newsletter-go/v3"
	"github.com/club-1/newsletter-go/v3/mailer"
	"github.com/club-1/newsletter-go/v3/mailer/mailertest"
	"github.com/club-1/newsletter-go/v3/messages"
	"github.com/mnako/letters"
)

func fakeNewsletter() *newsletter.Newsletter {
	return &newsletter.Newsletter{
		Config: &newsletter.Config{
			Dir: "/home/user/.config/newsletter",
			Emails: []string{
				"recipient@club1.fr",
			},
			Secret: "BASIC_SECRET",
			Settings: newsletter.Settings{
				Title:       "Title",
				DisplayName: "Display Name",
				Language:    messages.LangFrench,
			},
			Signature: "Bye bye",
		},
		Hostname:  "club1.fr",
		LocalUser: "user",
	}
}

func setupTest(t *testing.T, stdin string) *DummySyslog {
	t.Helper()

	var err error
	syslog := &DummySyslog{}
	logger = &Logger{Writer: syslog}
	t.Cleanup(func() { logger = nil })

	nl = fakeNewsletter()
	t.Cleanup(func() { nl = nil })

	incomingEmail, fromAddr, err = parseEmail(strings.NewReader(stdin))
	if err != nil {
		t.Fatalf("parse email: %v", err)
	}
	t.Cleanup(func() {
		incomingEmail = letters.Email{}
		fromAddr = ""
	})

	return syslog
}

func TestSubscribe(t *testing.T) {
	stdin := `From: test@club1.fr
To: user+subscribe@club1.fr
Message-Id: <fakeid@club1.fr>
Subject: Subscribe

`
	syslog := setupTest(t, stdin)

	var mail *mailer.Mail
	nl.Mailer = &mailertest.Mailer{Handler: func(m *mailer.Mail) error {
		mail = m
		return nil
	}}

	err := subscribe()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Log(syslog.String())

	expected := &mailer.Mail{
		FromAddr:        "user@club1.fr",
		FromName:        "Display Name",
		To:              "test@club1.fr",
		Id:              "<user-NRGABAKKE6AKVXM5S7IJQOUFFOXC2B3UF5QWX5VYFAKBRNWHZBHQ====@club1.fr>",
		InReplyTo:       "<fakeid@club1.fr>",
		ReplyTo:         "user+subscribe-confirm@club1.fr",
		ListUnsubscribe: "<mailto:user+unsubscribe@club1.fr>",
		Subject:         "[Title] Please confirm your subsciption",
		Body:            "Reply to this email to confirm that you want to subscribe to the newsletter [Title] (the content does not matter).\n\n-- \nBye bye",
	}
	if !reflect.DeepEqual(mail, expected) {
		t.Errorf("expected mail:\n%#v\ngot:\n%#v", expected, mail)
	}
}
