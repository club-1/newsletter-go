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
	"log/slog"
	"strings"
	"testing"
)

type DummySyslog struct {
	buf strings.Builder
}

func (s *DummySyslog) write(prefix, message string) error {
	s.buf.WriteString(prefix)
	s.buf.WriteString(message)
	s.buf.WriteByte('\n')
	return nil
}

func (s *DummySyslog) Info(message string) error {
	return s.write("i: ", message)
}

func (s *DummySyslog) Warning(message string) error {
	return s.write("w: ", message)
}

func (s *DummySyslog) Err(message string) error {
	return s.write("e: ", message)
}

func (s *DummySyslog) String() string {
	return s.buf.String()
}

func TestSyslogHandler(t *testing.T) {
	syslog := &DummySyslog{}
	logger := slog.New(NewSyslogHandler(syslog))

	logger.Info("hello", "what", "world")
	logger.Warn("warning")
	logger.Error("error")
	logger2 := logger.With("key", "value")
	logger2.Info("hello again")
	logger.Info("no attrs")
	logger3 := logger.WithGroup("test")
	logger3.Info("msg", "key", "value")
	logger.Info("msg", "key", "value")
	logger4 := logger3.With("coucou", "loulou")
	logger4.Warn("hello")

	expected := `i: what=world hello
w: warning
e: error
i: key=value hello again
i: no attrs
i: test.key=value msg
i: key=value msg
w: test.coucou=loulou hello
`
	if syslog.String() != expected {
		t.Errorf("expected:\n%sgot:\n%s", expected, syslog.String())
	}
}
