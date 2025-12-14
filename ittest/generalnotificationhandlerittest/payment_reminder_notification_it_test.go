package ittest

import (
	"context"
	"testing"

	"github.com/domesama/chat-and-notifications/generalnotifications"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/stretchr/testify/suite"
)

type PaymentReminderNotificationITTestSuite struct {
	BaseGeneralNotificationHandlerITTestSuite
}

func (t *PaymentReminderNotificationITTestSuite) SetupSuite() {
	t.BaseGeneralNotificationHandlerITTestSuite.SetupSuite()
}

func TestPaymentReminderNotificationITTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentReminderNotificationITTestSuite))
}

func (t *PaymentReminderNotificationITTestSuite) TestPaymentReminderNotificationWebSocketCommunication() {
	ctx := context.Background()
	userID := "user789"

	paymentReminders := []model.PaymentReminder{
		stub.CreatePaymentReminderNotificationPayload("ORD-001", 15.99, 0, "Payment due soon"),
		stub.CreatePaymentReminderNotificationPayload("ORD-002", 29.99, 3, "Payment overdue by 3 days"),
	}

	msgChan := t.subscribeToNotificationWebSocket(ctx, userID)

	callNotificationForwardingAPI(
		t.T(),
		t.cnt.HTTPServer.GetRunningPort(),
		"/notifications/payment-reminder",
		userID,
		paymentReminders...,
	)

	doneAssertion := assertNotifications(
		t.T(),
		generalnotifications.PaymentReminderNotification,
		msgChan,
		paymentReminders...,
	)
	<-doneAssertion
}
