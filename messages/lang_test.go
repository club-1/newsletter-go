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

package messages_test

import (
	"testing"

	"github.com/club-1/newsletter-go/v3/messages"
)

func TestSetLanguage(t *testing.T) {
	cases := []struct {
		language string
		expected string
	}{
		{"en", "Already subscribed"},
		{"fr", "Déjà inscrit"},
		{"", "Already subscribed"},
		{"unknown", "Already subscribed"},
	}
	for _, c := range cases {
		t.Run(c.language, func(t *testing.T) {
			messages.SetLanguage(messages.Language(c.language))
			actual := messages.AlreadySubscribed_subject.Print()
			if actual != c.expected {
				t.Errorf("expected %q, got %q", c.expected, actual)
			}
		})
	}
}
