package newsletter_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/club-1/newsletter-go"
	"github.com/club-1/newsletter-go/messages"
)

func TestInitConfig(t *testing.T) {
	cases := []struct {
		name     string
		expected *newsletter.Config
	}{
		{
			"basic",
			&newsletter.Config{
				Emails: []string{},
				Secret: "BASIC_SECRET\n",
				Settings: newsletter.Settings{
					Title:       "Title",
					DisplayName: "Display Name",
					Language:    messages.LangFrench,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			subTestInitConfig(t, c.name, c.expected)
		})
	}
}

func subTestInitConfig(t *testing.T, name string, expected *newsletter.Config) {
	configDir, err := filepath.Abs("testdata/config_" + name)
	if err != nil {
		t.Fatal(err)
	}
	expected.Dir = configDir
	config, err := newsletter.InitConfig(configDir)
	if err != nil {
		t.Errorf("init config: unexpected error: %v", err)
	}
	if !reflect.DeepEqual(expected, config) {
		t.Errorf("expected:\n%#v\ngot:\n%#v", expected, config)
	}
}
