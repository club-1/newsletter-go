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

package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func assertFileMatch(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}
	match, err := regexp.Match(expected, content)
	if err != nil {
		t.Error(err)
		return
	}
	if !match {
		t.Errorf("expected match:\n%s\ngot:\n%s\n", expected, content)
	}

}

func TestInitForwardFiles(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	err := initForwardFiles()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedFiles := map[string]string{
		".forward+subscribe":         `^\| "/[\w/-]+/sbin/newsletterctl subscribe"\n$`,
		".forward+subscribe-confirm": `^\| "/[\w/-]+/sbin/newsletterctl subscribe-confirm"\n$`,
		".forward+unsubscribe":       `^\| "/[\w/-]+/sbin/newsletterctl unsubscribe"\n$`,
		".forward+send":              `^\| "/[\w/-]+/sbin/newsletterctl send"\n$`,
		".forward+send-confirm":      `^\| "/[\w/-]+/sbin/newsletterctl send-confirm"\n$`,
	}
	for file, expected := range expectedFiles {
		assertFileMatch(t, filepath.Join(homeDir, file), expected)
	}
}
