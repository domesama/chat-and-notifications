package email

import (
	"github.com/kelseyhightower/envconfig"
)

// EmailConfig contains SMTP server settings
type EmailConfig struct {
	SMTPHost     string `envconfig:"SMTP_HOST" required:"true"`
	SMTPPort     int    `envconfig:"SMTP_PORT" default:"587"`
	SMTPUsername string `envconfig:"SMTP_USERNAME" required:"true"`
	SMTPPassword string `envconfig:"SMTP_PASSWORD" required:"true"`
	FromAddress  string `envconfig:"EMAIL_FROM_ADDRESS" required:"true"`
	FromName     string `envconfig:"EMAIL_FROM_NAME"`
}

func ProvideEmailConfig() (conf EmailConfig) {
	envconfig.MustProcess("", &conf)
	return
}
