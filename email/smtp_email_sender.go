package email

import (
	"context"
	"fmt"

	"github.com/wneessen/go-mail"
)

// SMTPEmailSender implements EmailSender using go-mail SMTP client
type SMTPEmailSender struct {
	config     EmailConfig
	mailClient *mail.Client
}

// NewSMTPEmailSender creates a new SMTP-based email sender
// @@wire-name@@ name:"EmailSenderSet"
func ProvideSMTPEmailSender(conf EmailConfig) (mailSender EmailSender, cleanUp func(), err error) {

	mailClient, err := mail.NewClient(
		conf.SMTPHost,
		mail.WithPort(conf.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(conf.SMTPUsername),
		mail.WithPassword(conf.SMTPPassword),
	)
	if err != nil {
		return
	}

	mailSender = SMTPEmailSender{config: conf, mailClient: mailClient}
	cleanUp = func() {
		_ = mailClient.Close()
	}

	return
}

func (s SMTPEmailSender) SendEmail(ctx context.Context, email EmailMessage) error {
	msg, err := GenerateMailMessage(email)
	if err != nil {
		return fmt.Errorf("failed to generate mail message: %w", err)
	}

	return s.mailClient.DialAndSendWithContext(ctx, msg)
}

func GenerateMailMessage(email EmailMessage) (msg *mail.Msg, err error) {
	msg = mail.NewMsg()
	if err = msg.From(email.SenderMailAddress); err != nil {
		return
	}

	if err = msg.To(email.ReceiverMailAddress); err != nil {
		return
	}

	msg.Subject(fmt.Sprintf("New message from %s", email.Subject))

	emailBody := email.Body
	msg.SetBodyString(mail.TypeTextPlain, emailBody)

	return
}
