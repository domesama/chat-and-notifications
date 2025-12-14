package stub

import (
	"context"

	"github.com/domesama/chat-and-notifications/email"
)

func WithReceiverEmail(receiverEmail string) Predicate[email.EmailMessage] {
	return func(ctx context.Context, msg email.EmailMessage) bool {
		return msg.ReceiverMailAddress == receiverEmail
	}
}

func WithEmailSubject(subject string) Predicate[email.EmailMessage] {
	return func(ctx context.Context, msg email.EmailMessage) bool {
		return msg.Subject == subject
	}
}

func WithEmailBody(body string) Predicate[email.EmailMessage] {
	return func(ctx context.Context, msg email.EmailMessage) bool {
		return msg.Body == body
	}
}
