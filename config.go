package newsletter

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"
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

	EnvLogDir string = "NEWSLETTER_LOG_DIR"
)

var (
	Conf        *Config
	HomeDir     string
	ConfigPath  string = ".config/newsletter"
	LocalUser   string
	Hostname    string
	LogDir      string
	LogFilePath string

	Routes = [...]string{RouteSubscribe, RouteSubscribeConfirm, RouteUnSubscribe, RouteSend, RouteSendConfirm}
)

type Settings struct {
	Title       string
	DisplayName string
	Language    string
}

type Config struct {
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
	emailsFilePath := filepath.Join(HomeDir, ConfigPath, EmailsFile)
	err := writeLines(c.Emails, emailsFilePath)
	if err != nil {
		return fmt.Errorf("could not save emails: %w", err)
	}
	return nil
}

func (c *Config) SaveSignature() error {
	signatureFilePath := filepath.Join(HomeDir, ConfigPath, SignatureFile)
	err := os.WriteFile(signatureFilePath, []byte(c.Signature), 0660)
	if err != nil {
		return fmt.Errorf("could not save signature: %w", err)
	}
	return nil
}

func (c *Config) SaveSettings() error {
	settingsFilePath := filepath.Join(HomeDir, ConfigPath, SettingsFile)
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

// Logger dir use environnement var
// if not defined, it fallback to user cache directory
func InitLogger(name string) *os.File {
	LogDir = os.Getenv(EnvLogDir)
	if LogDir == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Fatalln("cannot get user cache directory:", err)
		}

		LogDir = filepath.Join(userCacheDir, "newsletter")
	}

	err := os.MkdirAll(LogDir, 0775)
	if err != nil {
		log.Fatalln("cannot create log folder:", err)
	}

	LogFilePath = filepath.Join(LogDir, name+".log")

	logFile, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	if err != nil {
		log.Fatalln("cannot create or read log file: %w", err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	return logFile
}

// Load config in newsletter.Conf struct
// also get username
func ReadConfig() error {
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		return fmt.Errorf("could not get hostname: %w", err)
	}

	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get local user: %w", err)
	}
	LocalUser = user.Username

	HomeDir, err = os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user's home: %w", err)
	}

	configDir := filepath.Join(HomeDir, ConfigPath)
	err = os.MkdirAll(configDir, 0775)
	if err != nil {
		return fmt.Errorf("could not init config directory: %w", err)
	}

	var emails []string
	emailsFilePath := filepath.Join(HomeDir, ConfigPath, EmailsFile)
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
	signatureFilePath := filepath.Join(HomeDir, ConfigPath, SignatureFile)
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
	secretFilePath := filepath.Join(HomeDir, ConfigPath, SecretFile)
	_, err = os.Stat(secretFilePath)
	if errors.Is(err, os.ErrNotExist) {
		secret = randString()
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
	settingsFilePath := filepath.Join(HomeDir, ConfigPath, SettingsFile)
	_, err = os.Stat(settingsFilePath)
	if errors.Is(err, os.ErrNotExist) {
		settings = Settings{}
		settingsJson, err := json.Marshal(settings)
		if err != nil {
			return fmt.Errorf("could not encode settings JSON: %w", err)
		}
		err = os.WriteFile(settingsFilePath, settingsJson, 0660)
		if err != nil {
			return fmt.Errorf("could not write settings: %w", err)
		}
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
