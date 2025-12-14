package email

import (
	"context"
)

// EmailSender is an interface for sending emails
type EmailSender interface {
	SendEmail(ctx context.Context, mail EmailMessage) error
}
