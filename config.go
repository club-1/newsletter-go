package newsletter

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

const (
	EmailsFile    string = "emails"
	SecretFile    string = ".secret"
	SignatureFile string = "signature.txt"
	SettingsFile  string = "settings.json"
)

var (
	Conf       *Config
	ConfigPath string = ".config/newsletter"
)

type Settings struct {
	Title       string
	DisplayName string
}

type Config struct {
	Emails    []string
	Secret    string
	Signature string
	Settings  Settings
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
// func writeLines(lines []string, path string) error {
// 	file, err := os.Create(path)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	w := bufio.NewWriter(file)
// 	for _, line := range lines {
// 		fmt.Fprintln(w, line)
// 	}
// 	return w.Flush()
// }

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Load config in newsletter.Conf struct
func ReadConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user's home: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigPath)
	err = os.MkdirAll(configDir, 0775)
	if err != nil {
		return fmt.Errorf("could not init config directory: %w", err)
	}

	var emails []string
	emailsFilePath := filepath.Join(homeDir, ConfigPath, EmailsFile)
	_, err = os.Stat(emailsFilePath)
	if errors.Is(err, os.ErrNotExist) {
		emails = []string{}
	} else {
		emails, err = readLines(emailsFilePath)
		if err != nil {
			return fmt.Errorf("could not get emails: %w", err)
		}
	}

	var signature string
	signatureFilePath := filepath.Join(homeDir, ConfigPath, SignatureFile)
	_, err = os.Stat(signatureFilePath)
	if errors.Is(err, os.ErrNotExist) {
		signature = ""
	} else {
		signatureB, err := os.ReadFile(signatureFilePath)
		if err != nil {
			return fmt.Errorf("could not get signature: %w", err)
		}
		signature = string(signatureB)
	}

	var secret string
	secretFilePath := filepath.Join(homeDir, ConfigPath, SecretFile)
	_, err = os.Stat(secretFilePath)
	if errors.Is(err, os.ErrNotExist) {
		secret = randStringRunes(40)
		err := os.WriteFile(secretFilePath, []byte(secret+"\n"), 0660)
		if err != nil {
			return fmt.Errorf("could not store generated secret: %w", err)
		}
		log.Print("generated secret")
	} else {
		secretB, err := os.ReadFile(secretFilePath)
		if err != nil {
			return fmt.Errorf("could not get secret: %w", err)
		}
		secret = string(secretB)
	}

	var settings Settings
	settingsFilePath := filepath.Join(homeDir, ConfigPath, SettingsFile)
	_, err = os.Stat(settingsFilePath)
	if errors.Is(err, os.ErrNotExist) {
		settings = Settings{}
	} else {
		settingsJson, err := os.ReadFile(settingsFilePath)
		if err != nil {
			return fmt.Errorf("could not get settings: %w", err)
		}
		err = json.Unmarshal(settingsJson, &settings)
		if err != nil {
			return fmt.Errorf("could not decode settings: %w", err)
		}
	}

	Conf = &Config{
		Emails:    emails,
		Signature: signature,
		Secret:    secret,
		Settings:  settings,
	}
	return nil
}
