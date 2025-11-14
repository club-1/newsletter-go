package newsletter

import (
	"bytes"
	"fmt"
	"mime/quotedprintable"
	"os/exec"
	"strings"
	"time"
)

const LocalServer = "club1.fr"

var LocalUser string

type Mail struct {
	FromAddr  string
	FromName  string
	To        string
	Id        string
	InReplyTo string
	ReplyTo   string
	Subject   string
	Body      string
}

func (m *Mail) From() string {
	return m.FromName + " <" + m.FromAddr + ">"
}

func PostmasterAddr() string {
	return "postmaster@" + LocalServer
}

func UnsubscribeAddr() string {
	return LocalUser + "+" + RouteUnSubscribe + "@" + LocalServer
}

func Brackets(addr string) string {
	return "<" + addr + ">"
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

func SendMail(mail *Mail) error {
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
		"-a", fmt.Sprintf("List-Unsubscribe: <mailto:%s>", UnsubscribeAddr()),
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

// pre-fill the base mail with default values
func DefaultMail(subject string, body string) *Mail {
	if Conf.Settings.Title != "" {
		subject = "[" + Conf.Settings.Title + "] " + subject
	}

	if Conf.Signature != "" {
		body = body + "\n\n-- \n" + Conf.Signature
	}

	return &Mail{
		FromAddr: LocalUser + "@" + LocalServer,
		FromName: Conf.Settings.DisplayName,
		Subject:  subject,
		Body:     body,
	}
}

// send the newsletter to all the subscribed addresses
func SendNews(mail *Mail) error {
	var errCount = 0
	for _, address := range Conf.Emails {
		mail.To = address
		err := SendMail(mail)
		if err != nil {
			errCount++
		}
		time.Sleep(200 * time.Millisecond)
	}
	if errCount > 0 {
		return fmt.Errorf("error occured while sending mail to %v addresses", errCount)
	}
	return nil
}
