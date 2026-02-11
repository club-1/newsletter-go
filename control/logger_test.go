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

func (s *DummySyslog) Crit(message string) error {
	return s.write("c: ", message)
}

func (s *DummySyslog) String() string {
	return s.buf.String()
}

func TestLogger(t *testing.T) {
	syslog := &DummySyslog{}
	logger := &Logger{Writer: syslog}

	logger.Infof("hello %s", "world")
	logger.Warningf("warning")
	logger.Errorf("error")
	logger.Criticalf("critical")
	logger.AddContext("context")
	logger.Infof("hello again")

	expected := `i: hello world
w: warning
e: error
c: critical
i: context: hello again
`
	if syslog.String() != expected {
		t.Errorf("expected:\n%qgot:\n%q", expected, syslog.String())
	}
}
