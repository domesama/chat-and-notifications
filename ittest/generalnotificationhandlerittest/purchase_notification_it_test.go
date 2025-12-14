package ittest

import (
	"context"
	"testing"

	"github.com/domesama/chat-and-notifications/generalnotifications"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/stretchr/testify/suite"
)

type PurchaseNotificationITTestSuite struct {
	BaseGeneralNotificationHandlerITTestSuite
}

func (t *PurchaseNotificationITTestSuite) SetupSuite() {
	t.BaseGeneralNotificationHandlerITTestSuite.SetupSuite()
}

func TestPurchaseNotificationITTestSuite(t *testing.T) {
	suite.Run(t, new(PurchaseNotificationITTestSuite))
}

func (t *PurchaseNotificationITTestSuite) TestPurchaseNotificationWebSocketCommunication() {
	ctx := context.Background()
	userID := "user123"

	purchasePayloads := []model.PurchaseUpdate{
		stub.CreatePurchaseNotificationPayload("ORD-001", "Premium Coffee", 15.99),
		stub.CreatePurchaseNotificationPayload("ORD-002", "Wireless Mouse", 29.99),
		stub.CreatePurchaseNotificationPayload("ORD-003", "Notebook Set", 12.50),
	}

	msgChan := t.subscribeToNotificationWebSocket(ctx, userID)

	callNotificationForwardingAPI(
		t.T(),
		t.cnt.HTTPServer.GetRunningPort(),
		"/notifications/purchase",
		userID,
		purchasePayloads...,
	)

	doneAssertion := assertNotifications(
		t.T(),
		generalnotifications.PurchaseNotification,
		msgChan,
		purchasePayloads...,
	)
	<-doneAssertion
}

func (t *PurchaseNotificationITTestSuite) TestMultipleDevicesReceivePurchaseNotification() {
	ctx := context.Background()
	userID := "user456"

	purchasePayload := stub.CreatePurchaseNotificationPayload("ORD-100", "Laptop", 999.99)

	// Simulate multiple devices for the same user
	msgChanDevice1 := t.subscribeToNotificationWebSocket(ctx, userID)
	msgChanDevice2 := t.subscribeToNotificationWebSocket(ctx, userID)
	msgChanDevice3 := t.subscribeToNotificationWebSocket(
		ctx,
		userID,
	)

	// Forward notification using generic helper
	callNotificationForwardingAPI(
		t.T(),
		t.cnt.HTTPServer.GetRunningPort(),
		"/notifications/purchase",
		userID,
		purchasePayload,
	)

	// All devices should receive the notification
	doneDevice1 := assertNotifications(
		t.T(),
		generalnotifications.PurchaseNotification,
		msgChanDevice1,
		purchasePayload,
	)
	doneDevice2 := assertNotifications(
		t.T(),
		generalnotifications.PurchaseNotification,
		msgChanDevice2,
		purchasePayload,
	)
	doneDevice3 := assertNotifications(
		t.T(),
		generalnotifications.PurchaseNotification,
		msgChanDevice3,
		purchasePayload,
	)

	<-doneDevice1
	<-doneDevice2
	<-doneDevice3
}
