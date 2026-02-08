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

package mailertest_test

import (
	"errors"
	"testing"

	"github.com/club-1/newsletter-go/v3/mailer"
	"github.com/club-1/newsletter-go/v3/mailer/mailertest"
)

func TestMailer(t *testing.T) {
	var expectedErr = errors.New("fake error")
	var expectedMail = &mailer.Mail{To: "test@club1.fr"}

	var mail *mailer.Mail
	handler := func(m *mailer.Mail) error {
		mail = m
		return expectedErr
	}

	mailer := &mailertest.Mailer{
		Handler: handler,
	}

	err := mailer.Send(expectedMail)

	if err != expectedErr {
		t.Errorf("expected error %#v, got %#v", expectedErr, err)
	}

	if mail != expectedMail {
		t.Errorf("expected mail %#v, got %#v", expectedMail, mail)
	}
}
