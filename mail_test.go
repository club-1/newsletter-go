package newsletter_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/club-1/newsletter-go"
)

func readConfig(t *testing.T, path string) {
	t.Helper()
	prevConfigPath := newsletter.ConfigPath
	t.Cleanup(func() { newsletter.ConfigPath = prevConfigPath })
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("read test config: get home dir: %v", err)
	}
	relPath, err := filepath.Rel(home, path)
	if err != nil {
		t.Fatalf("read test config: create path: %v", err)
	}
	newsletter.ConfigPath = relPath
	if err := newsletter.ReadConfig(); err != nil {
		t.Fatalf("read test config: %v", err)
	}
}

func TestSendMail(t *testing.T) {
	cases := []struct {
		name     string
		mail     *newsletter.Mail
		expected []string
	}{
		{
			"basic",
			&newsletter.Mail{
				FromAddr: "nouvelles@club1.fr",
				FromName: "Nouvelles de CLUB1",
				To:       "test@gmail.com",
				Subject:  "Le sujet",
			},
			[]string{
				`-s Le\\ sujet`,
				`-r Nouvelles\\ de\\ CLUB1\\ \\<nouvelles@club1.fr\\>`,
				// FIXME(nicolasp): I don't think this flag should always be set.
				// IMO it should be part of the *Mail struct to be able to set it or not.
				// `-a List-Unsubscribe:\\ \\<mailto:\w+\+unsubscribe@\w+\\>`,
				`-a Content-Transfer-Encoding:\\ quoted-printable`,
				`-a Content-Type:\\ text/plain\\;\\ charset=UTF-8`,
				`-- test@gmail.com`,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			subTestSendMail(t, c.mail, c.expected)
		})
	}
}

func subTestSendMail(t *testing.T, mail *newsletter.Mail, expected []string) {
	tmp := t.TempDir()
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(testdata, "bin")
	configPath := filepath.Join(testdata, "config_basic")
	mailxCmdOut := filepath.Join(tmp, "mailx_cmd")
	readConfig(t, configPath)
	t.Setenv("PATH", path)
	t.Setenv("MAILX_CMD_OUT", mailxCmdOut)

	if err := newsletter.SendMail(mail); err != nil {
		t.Errorf("call SendMail: %v", err)
	}

	mailxCmd, err := os.ReadFile(mailxCmdOut)
	if err != nil {
		t.Errorf("read mailx_cmd: %v", err)
	}

	for _, e := range expected {
		match, err := regexp.Match(e, mailxCmd)
		if err != nil {
			t.Fatalf("invalid regexp %q: %v", e, err)
		}
		if !match {
			t.Errorf("expected:\n%s\nto match:\n%s", mailxCmd, e)
		}
	}
}
