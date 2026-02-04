package mailer

import (
	"bytes"
	"fmt"
	"mime/quotedprintable"
	"os/exec"
	"strings"
)

type Mail struct {
	FromAddr        string
	FromName        string
	To              string
	Id              string
	InReplyTo       string
	ReplyTo         string
	ListUnsubscribe string
	Subject         string
	Body            string
}

func (m *Mail) From() string {
	return m.FromName + " <" + m.FromAddr + ">"
}

func quotedPrintable(s string) (string, error) {
	var ac bytes.Buffer
	w := quotedprintable.NewWriter(&ac)
	_, err := w.Write([]byte(s))
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	return ac.String(), nil
}

func Send(mail *Mail) error {
	if mail.To == "" {
		return fmt.Errorf("no recipient address found")
	}

	encodedBody, err := quotedPrintable(mail.Body)
	if err != nil {
		return fmt.Errorf("could not encode body: %w", err)
	}

	args := []string{
		"-s", mail.Subject,
		"-r", mail.From(),
		"-a", "Content-Transfer-Encoding: quoted-printable",
		"-a", "Content-Type: text/plain; charset=UTF-8",
	}
	if mail.Id != "" {
		args = append(args, "-a", "Message-Id: "+mail.Id)
	}
	if mail.InReplyTo != "" {
		args = append(args, "-a", "In-Reply-To: "+mail.InReplyTo)
	}
	if mail.ReplyTo != "" {
		args = append(args, "-a", "Reply-To: "+mail.ReplyTo)
	}
	if mail.ListUnsubscribe != "" {
		args = append(args, "-a", "List-Unsubscribe: "+mail.ListUnsubscribe)
	}
	args = append(args, "--", mail.To)

	cmd := exec.Command("mailx", args...)
	cmd.Stdin = strings.NewReader(encodedBody)
	var out strings.Builder
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not execute command: %w %s", err, out.String())
	}
	return nil
}
