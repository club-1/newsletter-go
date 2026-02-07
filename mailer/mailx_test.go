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

package mailer

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func setupMailx(t *testing.T) (cmdPath, stdinPath string) {
	t.Helper()
	tmp := t.TempDir()
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(testdata, "bin")
	cmdPath = filepath.Join(tmp, "mailx_cmd")
	stdinPath = filepath.Join(tmp, "mailx_stdin")
	t.Setenv("PATH", path)
	t.Setenv("MAILX_CMD", cmdPath)
	t.Setenv("MAILX_STDIN", stdinPath)
	return
}

func TestMailxFlags(t *testing.T) {
	cases := []struct {
		name     string
		mail     *Mail
		expected []string
	}{
		{
			"basic",
			&Mail{
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
			subTestMailxFlags(t, c.mail, c.expected)
		})
	}
}

func subTestMailxFlags(t *testing.T, mail *Mail, expected []string) {
	mailxCmdPath, _ := setupMailx(t)
	mailx := &mailxMailer{}

	if err := mailx.Send(mail); err != nil {
		t.Errorf("call SendMail: %v", err)
	}

	mailxCmd, err := os.ReadFile(mailxCmdPath)
	if err != nil {
		t.Errorf("read mailx cmd: %v", err)
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

func TestMailxBody(t *testing.T) {
	_, mailxStdinPath := setupMailx(t)
	mailx := &mailxMailer{}

	mail := &Mail{
		FromAddr: "nouvelles@club1.fr",
		FromName: "Nouvelles de CLUB1",
		To:       "test@gmail.com",
		Subject:  "Le sujet",
		Body:     "Coucou, ça dit quoi ?",
	}

	expected := []byte("Coucou, =C3=A7a dit quoi ?")

	if err := mailx.Send(mail); err != nil {
		t.Errorf("call SendMail: %v", err)
	}

	mailxStdin, err := os.ReadFile(mailxStdinPath)
	if err != nil {
		t.Errorf("read mailx stdin: %v", err)
	}
	if !bytes.Equal(mailxStdin, expected) {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, mailxStdin)
	}
}
