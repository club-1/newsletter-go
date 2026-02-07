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

type Mailer interface {
	Send(m *Mail) error
}

var defaultMailer Mailer = &mailxMailer{}

func Default() Mailer {
	return defaultMailer
}

// Send sends a mail using the default [Mailer].
//
// Deprecated: use [Default()] to get a usable [Mailer] instead.
func Send(mail *Mail) error {
	return defaultMailer.Send(mail)
}
