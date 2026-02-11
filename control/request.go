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
	"errors"
	"io"
	"net/mail"

	"github.com/mnako/letters"
)

// Request is an email received by the controller.
//
// It is a very basic wrapper around [letters.Email] that parses some
// additional header fields that we always want to be valid.
type Request struct {
	letters.Email
	From      *mail.Address
	MessageID string
}

func ParseRequest(r io.Reader) (*Request, error) {
	email, err := letters.ParseEmail(r)
	if err != nil {
		return nil, err
	}
	if len(email.Headers.From) == 0 {
		return nil, errors.New(`"From" field missing from header or empty`)
	}
	if email.Headers.MessageID == "" {
		return nil, errors.New(`"Message-ID" field missing from header or empty`)
	}
	return &Request{
		Email:     email,
		From:      email.Headers.From[0],
		MessageID: string(email.Headers.MessageID),
	}, nil
}
