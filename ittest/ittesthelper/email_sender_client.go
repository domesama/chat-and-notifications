package ittesthelper

import (
	"context"

	"github.com/domesama/chat-and-notifications/email"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/google/wire"
)

var _ email.EmailSender = (*SimpleEmailSender)(nil)

// SimpleEmailSenderSet provides an implementation of email.EmailSender for IT tests allowing us to simulate different behaviors based on stubs.
var SimpleEmailSenderSet = wire.NewSet(
	ProvideSimpleEmailSender,
	wire.Bind(new(email.EmailSender), new(*SimpleEmailSender)),
)

type SimpleEmailSender struct {
	Stubs []SimpleEmailSenderStub
}
type SimpleEmailSenderStub struct {
	stub.Predicates[email.EmailMessage]
	ExpectedError error
}

func ProvideSimpleEmailSender() *SimpleEmailSender {
	return &SimpleEmailSender{Stubs: make([]SimpleEmailSenderStub, 0)}
}

func (s *SimpleEmailSender) AddStub(stub ...SimpleEmailSenderStub) {
	s.Stubs = append(s.Stubs, stub...)
}

func (s *SimpleEmailSender) SendEmail(ctx context.Context, msg email.EmailMessage) error {

	for _, currentStub := range s.Stubs {
		if currentStub.IsSatisfied(ctx, msg) {
			return currentStub.ExpectedError
		}
	}

	return nil
}
