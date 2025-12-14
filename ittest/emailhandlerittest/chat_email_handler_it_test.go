package ittest

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/domesama/chat-and-notifications/emailhandler/service"
	"github.com/domesama/chat-and-notifications/ittest/ittesthelper"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/stretchr/testify/suite"
)

type ChatEmailHandlerITTestSuite struct {
	BaseEmailHandlerITTestSuite
}

func (t *ChatEmailHandlerITTestSuite) SetupSuite() {
	t.BaseEmailHandlerITTestSuite.SetupSuite()
}

func TestChatEmailHandlerITTestSuite(t *testing.T) {
	suite.Run(t, new(ChatEmailHandlerITTestSuite))
}

func (t *ChatEmailHandlerITTestSuite) callChatMailingServer(
	ctx context.Context,
	expectedStatusCode int,
	message ...model.ChatMessage,
) {
	for _, msg := range message {
		t.callEmailRoute(ctx, "/email/chat", expectedStatusCode, msg)
	}
}

func (t *ChatEmailHandlerITTestSuite) TestChatEmailSendingSuccess() {
	ctx := context.Background()

	t.insertEmailInfo(
		ctx, service.EmailInfo{
			UserID: "user-1",
			Email:  "user1@example.com",
			Name:   "User One",
		},
	)
	t.insertEmailInfo(
		ctx, service.EmailInfo{
			UserID: "user-2",
			Email:  "user2@example.com",
			Name:   "User Two",
		},
	)

	// Add stub for successful email sending with verification
	t.cnt.SimpleEmailSender.AddStub(
		ittesthelper.SimpleEmailSenderStub{
			Predicates: stub.NewPredicates(
				stub.WithReceiverEmail("user2@example.com"),
				stub.WithEmailSubject("You have a new chat message from User Two"),
				stub.WithEmailBody("Test message content"),
			),
			ExpectedError: nil,
		},
	)

	chatMessages := stub.CreateChatMessages(
		"user-1", "user-2",
		"Test message content",
	)

	t.callChatMailingServer(ctx, http.StatusCreated, chatMessages...)
}

func (t *ChatEmailHandlerITTestSuite) TestChatEmailSendingFailure() {
	ctx := context.Background()

	// Setup EmailInfo data
	t.insertEmailInfo(
		ctx, service.EmailInfo{
			UserID: "user-3",
			Email:  "user3@example.com",
			Name:   "User Three",
		},
	)
	t.insertEmailInfo(
		ctx, service.EmailInfo{
			UserID: "user-4",
			Email:  "user4@example.com",
			Name:   "User Four",
		},
	)

	// Add stub for email sending failure
	t.cnt.SimpleEmailSender.AddStub(
		ittesthelper.SimpleEmailSenderStub{
			Predicates: stub.NewPredicates(
				stub.WithReceiverEmail("user4@example.com"),
			),
			ExpectedError: errors.New("failed to send email"),
		},
	)

	chatMessages := stub.CreateChatMessages(
		"user-3", "user-4",
		"Another test message",
	)

	t.callChatMailingServer(ctx, http.StatusInternalServerError, chatMessages...)
}

func (t *ChatEmailHandlerITTestSuite) TestChatMissingEmailInfo() {
	ctx := context.Background()

	// Do not insert EmailInfo for receiver - simulating missing user email data
	chatMessages := stub.CreateChatMessages(
		"user-5", "user-6",
		"Message to non-existent user",
	)

	t.callChatMailingServer(ctx, http.StatusInternalServerError, chatMessages...)
}
