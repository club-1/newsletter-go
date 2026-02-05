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
	"path/filepath"
	"reflect"
	"testing"

	"github.com/club-1/newsletter-go/v3"
	"github.com/club-1/newsletter-go/v3/messages"
)

func TestInitConfig(t *testing.T) {
	cases := []struct {
		name     string
		expected *newsletter.Config
	}{
		{
			"basic",
			&newsletter.Config{
				Emails: []string{},
				Secret: "BASIC_SECRET",
				Settings: newsletter.Settings{
					Title:       "Title",
					DisplayName: "Display Name",
					Language:    messages.LangFrench,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			subTestInitConfig(t, c.name, c.expected)
		})
	}
}

func subTestInitConfig(t *testing.T, name string, expected *newsletter.Config) {
	configDir, err := filepath.Abs("testdata/config_" + name)
	if err != nil {
		t.Fatal(err)
	}
	expected.Dir = configDir
	config, err := newsletter.InitConfig(configDir)
	if err != nil {
		t.Errorf("init config: unexpected error: %v", err)
	}
	if !reflect.DeepEqual(expected, config) {
		t.Errorf("expected:\n%#v\ngot:\n%#v", expected, config)
	}
}
