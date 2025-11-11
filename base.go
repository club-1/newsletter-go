package newsletter

import (
	"bufio"
	"encoding/json"
	"fmt"
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

// Load config in newsletter.Conf struct
func ReadConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user's home: %w", err)
	}

	emailsFilePath := filepath.Join(homeDir, ConfigPath, EmailsFile)
	emails, err := readLines(emailsFilePath)
	if err != nil {
		return fmt.Errorf("could not get emails: %w", err)
	}
	signatureFilePath := filepath.Join(homeDir, ConfigPath, SignatureFile)
	signature, err := os.ReadFile(signatureFilePath)
	if err != nil {
		return fmt.Errorf("could not get signature: %w", err)
	}
	secretFilePath := filepath.Join(homeDir, ConfigPath, SecretFile)
	secret, err := os.ReadFile(secretFilePath)
	if err != nil {
		return fmt.Errorf("could not get secret: %w", err)
	}

	settingsFilePath := filepath.Join(homeDir, ConfigPath, SettingsFile)
	settingsJson, err := os.ReadFile(settingsFilePath)
	if err != nil {
		return fmt.Errorf("could not get settings: %w", err)
	}
	var settings Settings
	err = json.Unmarshal(settingsJson, &settings)
	if err != nil {
		return fmt.Errorf("could not decode settings: %w", err)
	}

	Conf = &Config{
		Emails:    emails,
		Signature: string(signature),
		Secret:    string(secret),
		Settings:  settings,
	}
	return nil
}
