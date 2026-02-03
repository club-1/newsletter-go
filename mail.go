package newsletter

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"github.com/club-1/newsletter-go/mailx"
)

func PostmasterAddr() string {
	return "postmaster@" + Hostname
}

func LocalUserAddr() string {
	return LocalUser + "@" + Hostname
}

func UnsubscribeAddr() string {
	return LocalUser + "+" + RouteUnSubscribe + "@" + Hostname
}

func SubscribeConfirmAddr() string {
	return LocalUser + "+" + RouteSubscribeConfirm + "@" + Hostname
}

func SendConfirmAddr() string {
	return LocalUser + "+" + RouteSendConfirm + "@" + Hostname
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base32.StdEncoding.EncodeToString(sum[0:32])
}

func HashWithSecret(s string) string {
	return hashString(s + Conf.Secret)
}

// generate a Message-ID
// it's based on incoming mail From address and local .secret file content
func GenerateId(hash string) string {
	return LocalUser + "-" + hash + "@" + Hostname
}

// retrive hash from message-ID using the form: `USER-HASH@SERVER`
func GetHashFromId(messageId string) (string, error) {
	after, prefixFound := strings.CutPrefix(messageId, LocalUser+"-")
	before, suffixFound := strings.CutSuffix(after, "@"+Hostname)
	if !prefixFound || !suffixFound {
		return "", fmt.Errorf("message ID does'nt match generated ID form")
	}
	return before, nil
}

func Brackets(addr string) string {
	return "<" + addr + ">"
}

// pre-fill the base mail with default values
func DefaultMail(subject string, body string) *mailx.Mail {
	if Conf.Settings.Title != "" {
		subject = "[" + Conf.Settings.Title + "] " + subject
	}

	if Conf.Signature != "" {
		body = body + "\n\n-- \n" + Conf.Signature
	}

	return &mailx.Mail{
		FromAddr:        LocalUser + "@" + Hostname,
		FromName:        Conf.Settings.DisplayName,
		ListUnsubscribe: fmt.Sprintf("<mailto:%s>", UnsubscribeAddr()),
		Subject:         subject,
		Body:            body,
	}
}

// add a `(preview)` text after original subject
func SendPreviewMail(mail *mailx.Mail) error {
	mail.To = LocalUserAddr()
	mail.Subject += " (preview)"

	err := mailx.Send(mail)
	if err != nil {
		return fmt.Errorf("could not send preview mail: %w", err)
	}
	fmt.Printf("📨 preview email send to %s\n", LocalUserAddr())
	return nil
}

// send the newsletter to all the subscribed addresses
func SendNews(mail *mailx.Mail) error {
	var errCount = 0
	for _, address := range Conf.Emails {
		mail.To = address
		err := mailx.Send(mail)
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
