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

package mailer_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/club-1/newsletter-go/v3/mailer"
)

func TestSend(t *testing.T) {
	cases := []struct {
		name     string
		mail     *mailer.Mail
		expected []string
	}{
		{
			"basic",
			&mailer.Mail{
				FromAddr: "nouvelles@club1.fr",
				FromName: "Nouvelles de CLUB1",
				To:       "test@gmail.com",
				Subject:  "Le sujet",
			},
			[]string{
				`-s Le\\ sujet`,
				`-r Nouvelles\\ de\\ CLUB1\\ \\<nouvelles@club1.fr\\>`,
				`-a Content-Transfer-Encoding:\\ quoted-printable`,
				`-a Content-Type:\\ text/plain\\;\\ charset=UTF-8`,
				`-- test@gmail.com`,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			subTestSend(t, c.mail, c.expected)
		})
	}
}

func subTestSend(t *testing.T, mail *mailer.Mail, expected []string) {
	tmp := t.TempDir()
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(testdata, "bin")
	mailxCmdOut := filepath.Join(tmp, "mailx_cmd")
	t.Setenv("PATH", path)
	t.Setenv("MAILX_CMD_OUT", mailxCmdOut)

	if err := mailer.Send(mail); err != nil {
		t.Errorf("call SendMail: %v", err)
	}

	mailxCmd, err := os.ReadFile(mailxCmdOut)
	if err != nil {
		t.Errorf("read mailx_cmd: %v", err)
	}

	for _, e := range expected {
		match, err := regexp.Match(e, mailxCmd)
		if err != nil {
			t.Fatalf("invalid regexp %q: %v", e, err)
		}
		if !match {
			t.Errorf("expected:\n%s\nto match:\n%s", mailxCmd, e)
		}
	}
}
