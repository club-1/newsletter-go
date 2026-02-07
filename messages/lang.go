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

package messages

type Language string

const (
	LangEnglish Language = "en"
	LangFrench  Language = "fr"
)

var language Language

func SetLanguage(l Language) {
	language = l
}

type Message struct {
	en string
	fr string
}

func (m Message) Print() string {
	switch language {
	case LangEnglish:
		return m.en
	case LangFrench:
		return m.fr
	default:
		return m.en
	}
}
