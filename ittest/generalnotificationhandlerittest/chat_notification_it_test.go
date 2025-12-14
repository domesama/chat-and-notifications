package ittest

import (
	"context"
	"testing"

	"github.com/domesama/chat-and-notifications/generalnotifications"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/stretchr/testify/suite"
)

type ChatNotificationITTestSuite struct {
	BaseGeneralNotificationHandlerITTestSuite
}

func (t *ChatNotificationITTestSuite) SetupSuite() {
	t.BaseGeneralNotificationHandlerITTestSuite.SetupSuite()
}

func TestChatNotificationITTestSuite(t *testing.T) {
	suite.Run(t, new(ChatNotificationITTestSuite))
}

func (t *ChatNotificationITTestSuite) TestChatNotification() {
	ctx := context.Background()
	userID := "Mr.A"

	chatMessages := stub.CreateChatMessages("Mr.A", "Mr.B", "Hello", "How are you?", "Goodbye")

	msgChan := t.subscribeToNotificationWebSocket(ctx, userID)

	callNotificationForwardingAPI(
		t.T(),
		t.cnt.HTTPServer.GetRunningPort(),
		"/notifications/chat",
		userID,
		chatMessages...,
	)

	doneAssertion := assertNotifications(
		t.T(),
		generalnotifications.ChatNotification,
		msgChan,
		chatMessages...,
	)
	<-doneAssertion
}
