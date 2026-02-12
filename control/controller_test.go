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
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/club-1/newsletter-go/v3"
	"github.com/club-1/newsletter-go/v3/mailer"
	"github.com/club-1/newsletter-go/v3/mailer/mailertest"
	"github.com/club-1/newsletter-go/v3/messages"
)

func fakeNewsletter(t *testing.T) *newsletter.Newsletter {
	t.Helper()
	return &newsletter.Newsletter{
		Config: &newsletter.Config{
			Dir: t.TempDir(),
			Emails: []string{
				"recipient@club1.fr",
			},
			Secret: "BASIC_SECRET",
			Settings: newsletter.Settings{
				Title:       "Title",
				DisplayName: "Display Name",
				Language:    messages.LangEnglish,
			},
			Signature: "Bye bye",
		},
		Hostname:  "club1.fr",
		LocalUser: "user",
	}
}

func setupTest(t *testing.T) (*Controller, *DummySyslog) {
	t.Helper()
	syslog := &DummySyslog{}
	logger := &Logger{Writer: syslog}
	nl := fakeNewsletter(t)
	messages.SetLanguage(nl.Config.Settings.Language)
	return &Controller{log: logger, nl: nl}, syslog
}

func handle(t *testing.T, route, stdin string) (*Controller, *DummySyslog, *mailer.Mail, error) {
	controller, syslog := setupTest(t)

	var mail *mailer.Mail
	controller.nl.Mailer = &mailertest.Mailer{Handler: func(m *mailer.Mail) error {
		mail = m
		return nil
	}}

	err := controller.Handle(route, strings.NewReader(stdin))
	return controller, syslog, mail, err
}

func TestSubscribe(t *testing.T) {
	route := newsletter.RouteSubscribe
	stdin := `From: test@club1.fr
To: user+subscribe@club1.fr
Message-Id: <fakeid@club1.fr>
Subject: Subscribe

`
	_, syslog, mail, err := handle(t, route, stdin)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// TODO: check syslog instead of just printing
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

func TestSubscribeConfirm(t *testing.T) {
	route := newsletter.RouteSubscribeConfirm
	stdin := `From: test@club1.fr
To: user+subscribe-confirm@club1.fr
Message-Id: <fakeid2@club1.fr>
In-Reply-To: <user-NRGABAKKE6AKVXM5S7IJQOUFFOXC2B3UF5QWX5VYFAKBRNWHZBHQ====@club1.fr>
Subject: Subscribe confirm

`
	c, syslog, mail, err := handle(t, route, stdin)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// TODO: check syslog instead of just printing
	t.Log(syslog.String())

	expectedAddr := "test@club1.fr"
	if !slices.Contains(c.nl.Config.Emails, expectedAddr) {
		t.Errorf("expected %q to be subscribed, got %v", expectedAddr, c.nl.Config.Emails)
	}

	expectedMail := &mailer.Mail{
		FromAddr:        "user@club1.fr",
		FromName:        "Display Name",
		To:              "test@club1.fr",
		InReplyTo:       "<fakeid2@club1.fr>",
		ListUnsubscribe: "<mailto:user+unsubscribe@club1.fr>",
		Subject:         "[Title] Subscription is successfull !",
		Body:            "Your email has been successfully subscribed to the newsletter [Title].\n\n-- \nBye bye",
	}

	if !reflect.DeepEqual(mail, expectedMail) {
		t.Errorf("expected mail:\n%#v\ngot:\n%#v", expectedMail, mail)
	}
}

func TestUnsubscribe(t *testing.T) {
	route := newsletter.RouteUnSubscribe
	stdin := `From: recipient@club1.fr
To: user+unsubscrib@club1.fr
Message-Id: <fakeid@club1.fr>
Subject: Unsubscribe

`
	c, syslog, mail, err := handle(t, route, stdin)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// TODO: check syslog instead of just printing
	t.Log(syslog.String())

	expectedAddr := "recipent@club1.fr"
	if slices.Contains(c.nl.Config.Emails, expectedAddr) {
		t.Errorf("expected %q to be unsubscribed, got %v", expectedAddr, c.nl.Config.Emails)
	}

	expectedMail := &mailer.Mail{
		FromAddr:        "user@club1.fr",
		FromName:        "Display Name",
		To:              "recipient@club1.fr",
		InReplyTo:       "<fakeid@club1.fr>",
		ListUnsubscribe: "<mailto:user+unsubscribe@club1.fr>",
		Subject:         "[Title] Unsubscription is successfull",
		Body:            "Your email has been successfully unsubscribed from the newsletter [Title].\n\n-- \nBye bye",
	}

	if !reflect.DeepEqual(mail, expectedMail) {
		t.Errorf("expected mail:\n%#v\ngot:\n%#v", expectedMail, mail)
	}
}
