package mailx_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/club-1/newsletter-go/mailx"
)

func TestSend(t *testing.T) {
	cases := []struct {
		name     string
		mail     *mailx.Mail
		expected []string
	}{
		{
			"basic",
			&mailx.Mail{
				FromAddr: "nouvelles@club1.fr",
				FromName: "Nouvelles de CLUB1",
				To:       "test@gmail.com",
				Subject:  "Le sujet",
			},
			[]string{
				`-s Le\\ sujet`,
				`-r Nouvelles\\ de\\ CLUB1\\ \\<nouvelles@club1.fr\\>`,
				`-a Content-Transfer-Encoding:\\ quoted-printable`,
				`-a Content-Type:\\ text/plain\\;\\ charset=UTF-8`,
				`-- test@gmail.com`,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			subTestSend(t, c.mail, c.expected)
		})
	}
}

func subTestSend(t *testing.T, mail *mailx.Mail, expected []string) {
	tmp := t.TempDir()
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(testdata, "bin")
	mailxCmdOut := filepath.Join(tmp, "mailx_cmd")
	t.Setenv("PATH", path)
	t.Setenv("MAILX_CMD_OUT", mailxCmdOut)

	if err := mailx.Send(mail); err != nil {
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
