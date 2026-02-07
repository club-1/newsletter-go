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
	"fmt"
	"mime/quotedprintable"
	"os/exec"
	"strings"
)

type Mail struct {
	FromAddr        string
	FromName        string
	To              string
	Id              string
	InReplyTo       string
	ReplyTo         string
	ListUnsubscribe string
	Subject         string
	Body            string
}

func (m *Mail) From() string {
	return m.FromName + " <" + m.FromAddr + ">"
}

func quotedPrintable(s string) (string, error) {
	var ac bytes.Buffer
	w := quotedprintable.NewWriter(&ac)
	_, err := w.Write([]byte(s))
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	return ac.String(), nil
}

func Send(mail *Mail) error {
	if mail.To == "" {
		return fmt.Errorf("no recipient address found")
	}

	encodedBody, err := quotedPrintable(mail.Body)
	if err != nil {
		return fmt.Errorf("encode body: %w", err)
	}

	args := []string{
		"-s", mail.Subject,
		"-r", mail.From(),
		"-a", "Content-Transfer-Encoding: quoted-printable",
		"-a", "Content-Type: text/plain; charset=UTF-8",
	}
	if mail.Id != "" {
		args = append(args, "-a", "Message-Id: "+mail.Id)
	}
	if mail.InReplyTo != "" {
		args = append(args, "-a", "In-Reply-To: "+mail.InReplyTo)
	}
	if mail.ReplyTo != "" {
		args = append(args, "-a", "Reply-To: "+mail.ReplyTo)
	}
	if mail.ListUnsubscribe != "" {
		args = append(args, "-a", "List-Unsubscribe: "+mail.ListUnsubscribe)
	}
	args = append(args, "--", mail.To)

	cmd := exec.Command("mailx", args...)
	cmd.Stdin = strings.NewReader(encodedBody)
	var out strings.Builder
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("execute command: %w %s", err, out.String())
	}
	return nil
}
