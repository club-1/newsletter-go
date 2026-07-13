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
	"context"
	"log/slog"
)

// Writer is the minimal interface required by [Logger] for its underlying writer.
type Writer interface {
	Info(message string) error
	Warning(message string) error
	Err(message string) error
}

// SyslogHandler is a basic shim between [slog.Handler] and [syslog.Writer].
type SyslogHandler struct {
	writer Writer
	group  []byte
	prefix []byte
}

// NewSyslogHandler creates a new [SyslogHandler] using the given writer.
func NewSyslogHandler(writer Writer) *SyslogHandler {
	return &SyslogHandler{writer: writer}
}

// Enabled reports whether the handler handles records at the given level.
func (h *SyslogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= slog.LevelInfo
}

// Handle handles the Record.
func (h *SyslogHandler) Handle(ctx context.Context, r slog.Record) (err error) {
	buf := make([]byte, 0, 1024)
	buf = append(buf, h.prefix...)
	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, a)
		return true
	})
	buf = append(buf, ' ')
	buf = append(buf, []byte(r.Message)...)
	buf = buf[1:] // Skip first byte which will always be a space.
	switch r.Level {
	case slog.LevelInfo:
		err = h.writer.Info(string(buf))
	case slog.LevelWarn:
		err = h.writer.Warning(string(buf))
	case slog.LevelError:
		err = h.writer.Err(string(buf))
	}
	return
}

func (h *SyslogHandler) appendAttr(buf []byte, a slog.Attr) []byte {
	buf = append(buf, ' ')
	buf = append(buf, h.group...)
	buf = append(buf, []byte(a.String())...)
	return buf
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
func (h *SyslogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var suffix []byte
	for _, a := range attrs {
		suffix = h.appendAttr(suffix, a)
	}
	h2 := *h
	h2.prefix = make([]byte, len(h.prefix), len(h.prefix)+len(suffix))
	copy(h2.prefix, h.prefix)
	h2.prefix = append(h2.prefix, suffix...)
	return &h2
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *SyslogHandler) WithGroup(name string) slog.Handler {
	h2 := *h
	h2.group = make([]byte, len(h.group), len(h.group)+len(name)+1)
	copy(h2.group, h.group)
	h2.group = append(h2.group, []byte(name)...)
	h2.group = append(h2.group, '.')
	return &h2
}
