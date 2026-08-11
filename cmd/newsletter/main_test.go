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
	"errors"
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

func assertFileNotExist(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		t.Errorf("file %s exists", path)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInitStop(t *testing.T) {
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

	err = stop(nil)
	for file := range expectedFiles {
		assertFileNotExist(t, filepath.Join(homeDir, file))
	}

}

func must(t *testing.T, f func() error) {
	t.Helper()
	if err := f(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSubjectBody(t *testing.T) {
	tmp := t.TempDir()
	stdin := filepath.Join(tmp, "stdin")
	body := filepath.Join(tmp, "body")
	cases := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		args            []string
		setStdin        bool
		expectedSubject string
		expectedBody    string
		expectError     bool
	}{
		{
			name:        "empty args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "single arg without stdin",
			args:        []string{"subject"},
			setStdin:    false,
			expectError: true,
		},
		{
			name:            "single arg with stdin",
			args:            []string{"subject"},
			setStdin:        true,
			expectedSubject: "subject",
			expectedBody:    "stdin",
		},
		{
			name:        "two args non-existing body file",
			args:        []string{"subject", "non-existing"},
			expectError: true,
		},
		{
			name:            "two args existing body file",
			args:            []string{"subject", body},
			expectedSubject: "subject",
			expectedBody:    "body",
		},
		{
			name:        "three args",
			args:        []string{"one", "two", "three"},
			expectError: true,
		},
	}
	must(t, func() error { return os.WriteFile(stdin, []byte("stdin"), 0o666) })
	must(t, func() error { return os.WriteFile(body, []byte("body"), 0o666) })
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.setStdin {
				var err error
				prevStdin := os.Stdin
				t.Cleanup(func() { os.Stdin = prevStdin })
				os.Stdin, err = os.Open(stdin)
				if err != nil {
					t.Fatal(err)
				}

			}
			subject, body, err := getSubjectBody(c.args)
			if err != nil {
				if !c.expectError {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if c.expectError {
				t.Fatal("expected an error")
			}
			if subject != c.expectedSubject {
				t.Errorf("subject = %v, expected %v", subject, c.expectedSubject)
			}
			if body != c.expectedBody {
				t.Errorf("body = %v, expected %v", body, c.expectedBody)
			}
		})
	}
}
