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

package newsletter

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/club-1/newsletter-go/messages"
)

const (
	EmailsFile    string = "emails"
	SecretFile    string = ".secret"
	SignatureFile string = "signature.txt"
	SettingsFile  string = "settings.json"

	RouteSubscribe        string = "subscribe"
	RouteSubscribeConfirm string = "subscribe-confirm"
	RouteUnSubscribe      string = "unsubscribe"
	RouteSend             string = "send"
	RouteSendConfirm      string = "send-confirm"

	LogIdentifier string = "newsletter"
)

var (
	Routes = [...]string{RouteSubscribe, RouteSubscribeConfirm, RouteUnSubscribe, RouteSend, RouteSendConfirm}
)

type Settings struct {
	Title       string
	DisplayName string
	Language    messages.Language
}

type Config struct {
	Dir       string
	Emails    []string
	Secret    string
	Signature string
	Settings  Settings
}

func (c *Config) Unsubscribe(addr string) error {
	index := slices.Index(c.Emails, addr)
	if index == -1 {
		return fmt.Errorf("not subscribed")
	}
	c.Emails = append(c.Emails[:index], c.Emails[index+1:]...)
	return c.saveEmails()
}

func (c *Config) Subscribe(addr string) error {
	c.Emails = append(c.Emails, addr)
	return c.saveEmails()
}

func (c *Config) saveEmails() error {
	emailsFilePath := filepath.Join(c.Dir, EmailsFile)
	err := writeLines(c.Emails, emailsFilePath)
	if err != nil {
		return fmt.Errorf("could not save emails: %w", err)
	}
	return nil
}

func (c *Config) SaveSignature() error {
	signatureFilePath := filepath.Join(c.Dir, SignatureFile)
	err := os.WriteFile(signatureFilePath, []byte(c.Signature), 0660)
	if err != nil {
		return fmt.Errorf("could not save signature: %w", err)
	}
	return nil
}

func (c *Config) SaveSettings() error {
	settingsFilePath := filepath.Join(c.Dir, SettingsFile)
	settingsJson, err := json.Marshal(c.Settings)
	if err != nil {
		return fmt.Errorf("could not encode settings JSON: %w", err)
	}
	err = os.WriteFile(settingsFilePath, settingsJson, 0660)
	if err != nil {
		return fmt.Errorf("could not write settings: %w", err)
	}
	return nil
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	content := strings.Join(lines, "\n")
	err := os.WriteFile(path, []byte(content+"\n"), 0664)
	if err != nil {
		return fmt.Errorf("write file error: %w", err)
	}
	return nil
}

func randString() string {
	key := make([]byte, 32)
	rand.Read(key)
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(key)))
	base64.StdEncoding.Encode(dst, key)
	return string(dst)
}

// InitConfig returns a new [*Config] loaded from the given configDir.
func InitConfig(configDir string) (*Config, error) {
	err := os.MkdirAll(configDir, 0775)
	if err != nil {
		return nil, fmt.Errorf("could not init config directory: %w", err)
	}

	var emails []string
	emailsFilePath := filepath.Join(configDir, EmailsFile)
	_, err = os.Stat(emailsFilePath)
	if errors.Is(err, os.ErrNotExist) {
		emails = []string{}
	} else {
		emails, err = readLines(emailsFilePath)
		if err != nil {
			return nil, fmt.Errorf("could not get emails: %w", err)
		}
	}

	var signature string
	signatureFilePath := filepath.Join(configDir, SignatureFile)
	_, err = os.Stat(signatureFilePath)
	if errors.Is(err, os.ErrNotExist) {
		signature = ""
	} else {
		signatureB, err := os.ReadFile(signatureFilePath)
		if err != nil {
			return nil, fmt.Errorf("could not get signature: %w", err)
		}
		signature = string(signatureB)
	}

	var secret string
	secretFilePath := filepath.Join(configDir, SecretFile)
	_, err = os.Stat(secretFilePath)
	if errors.Is(err, os.ErrNotExist) {
		secret = randString()
		err := os.WriteFile(secretFilePath, []byte(secret+"\n"), 0660)
		if err != nil {
			return nil, fmt.Errorf("could not store generated secret: %w", err)
		}
		log.Print("generated secret")
	} else {
		secretB, err := os.ReadFile(secretFilePath)
		if err != nil {
			return nil, fmt.Errorf("could not get secret: %w", err)
		}
		secret = string(secretB)
	}

	var settings Settings
	settingsFilePath := filepath.Join(configDir, SettingsFile)
	_, err = os.Stat(settingsFilePath)
	if errors.Is(err, os.ErrNotExist) {
		settings = Settings{}
		settingsJson, err := json.Marshal(settings)
		if err != nil {
			return nil, fmt.Errorf("could not encode settings JSON: %w", err)
		}
		err = os.WriteFile(settingsFilePath, settingsJson, 0660)
		if err != nil {
			return nil, fmt.Errorf("could not write settings: %w", err)
		}
	} else {
		settingsJson, err := os.ReadFile(settingsFilePath)
		if err != nil {
			return nil, fmt.Errorf("could not get settings: %w", err)
		}
		err = json.Unmarshal(settingsJson, &settings)
		if err != nil {
			return nil, fmt.Errorf("could not decode settings: %w", err)
		}
	}

	return &Config{
		Dir:       configDir,
		Emails:    emails,
		Signature: signature,
		Secret:    secret,
		Settings:  settings,
	}, nil
}
