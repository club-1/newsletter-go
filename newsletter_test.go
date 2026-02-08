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

package newsletter_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/club-1/newsletter-go/v3"
	"github.com/club-1/newsletter-go/v3/mailer"
	"github.com/club-1/newsletter-go/v3/mailer/mailertest"
	"github.com/club-1/newsletter-go/v3/messages"
)

func TestNew(t *testing.T) {
	homeDir, err := filepath.Abs("testdata/home")
	if err != nil {
		t.Fatalf("get fake home path: %v", err)
	}
	os.Setenv("HOME", homeDir)
	nl, err := newsletter.New()
	if err != nil {
		t.Errorf("new: unexpected error: %v", err)
	}
	expectedConfig := &newsletter.Config{
		Dir:    filepath.Join(homeDir, newsletter.ConfigPath),
		Emails: []string{},
		Secret: "BASIC_SECRET",
		Settings: newsletter.Settings{
			Title:       "Title",
			DisplayName: "Display Name",
			Language:    messages.LangFrench,
		},
	}
	if !reflect.DeepEqual(nl.Config, expectedConfig) {
		t.Errorf("expected config:\n%#v\ngot:\n%#v", expectedConfig, nl.Config)
	}
	if nl.Hostname == "" {
		t.Errorf("expected non empty Hostname")
	}
	if nl.LocalUser == "" {
		t.Errorf("expected non empty LocalUser")
	}
}

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

func TestDefaultMail(t *testing.T) {
	nl := fakeNewsletter()
	mail := nl.DefaultMail("Test subject", "Mail body")
	expected := &mailer.Mail{
		FromAddr:        "user@club1.fr",
		FromName:        "Display Name",
		ListUnsubscribe: "<mailto:user+unsubscribe@club1.fr>",
		Subject:         "[Title] Test subject",
		Body: `Mail body

-- 
Bye bye`,
	}
	if !reflect.DeepEqual(mail, expected) {
		t.Errorf("expected:\n%#v\ngot:\n%#v", expected, mail)
	}
}

func TestSendPreviewMail(t *testing.T) {
	var actual *mailer.Mail
	nl := fakeNewsletter()
	nl.Mailer = &mailertest.Mailer{Handler: func(mail *mailer.Mail) error {
		actual = mail
		return nil
	}}

	mail := &mailer.Mail{
		FromAddr: "user@club1.fr",
		Subject:  "Coucou les loulous",
	}
	expected := &mailer.Mail{
		FromAddr: "user@club1.fr",
		Subject:  "Coucou les loulous (preview)",
		To:       "user@club1.fr",
	}

	err := nl.SendPreviewMail(mail)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%#v\ngot:\n%#v", expected, actual)
	}
}
