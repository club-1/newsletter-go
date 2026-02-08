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
	"fmt"
	"strings"
)

// Writer is the minimal interface required by [Logger] for its underlying writer.
type Writer interface {
	Info(message string) error
	Warning(message string) error
	Err(message string) error
	Crit(message string) error
}

// Logger is a basic wrapper around [syslog.Writer] that allows to add context
// and offers formatting methods.
type Logger struct {
	Writer Writer
	ctx    strings.Builder
}

func (l *Logger) AddContext(v string) {
	l.ctx.WriteString(v)
	l.ctx.WriteString(": ")
}

func (l *Logger) Infof(format string, v ...any) error {
	return l.Writer.Info(l.ctx.String() + fmt.Sprintf(format, v...))
}

func (l *Logger) Warningf(format string, v ...any) error {
	return l.Writer.Warning(l.ctx.String() + fmt.Sprintf(format, v...))
}

func (l *Logger) Errorf(format string, v ...any) error {
	return l.Writer.Err(l.ctx.String() + fmt.Sprintf(format, v...))
}

func (l *Logger) Criticalf(format string, v ...any) error {
	return l.Writer.Crit(l.ctx.String() + fmt.Sprintf(format, v...))
}
