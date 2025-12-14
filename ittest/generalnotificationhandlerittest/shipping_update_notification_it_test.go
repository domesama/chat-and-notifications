package ittest

import (
	"context"
	"testing"

	"github.com/domesama/chat-and-notifications/generalnotifications"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/stretchr/testify/suite"
)

type ShippingUpdateNotificationITTestSuite struct {
	BaseGeneralNotificationHandlerITTestSuite
}

func (t *ShippingUpdateNotificationITTestSuite) SetupSuite() {
	t.BaseGeneralNotificationHandlerITTestSuite.SetupSuite()
}

func TestShippingUpdateNotificationITTestSuite(t *testing.T) {
	suite.Run(t, new(ShippingUpdateNotificationITTestSuite))
}

func (t *ShippingUpdateNotificationITTestSuite) TestShippingUpdateNotificationWebSocketCommunication() {
	ctx := context.Background()
	userID := "user999"

	shippingUpdates := []model.ShippingUpdate{
		stub.CreateShippingUpdateNotificationPayload(
			"ORD-001",
			"TRK123456",
			"in_transit",
			"New York Distribution Center",
		),
		stub.CreateShippingUpdateNotificationPayload("ORD-001", "TRK123456", "out_for_delivery", "Local Delivery Hub"),
		stub.CreateShippingUpdateNotificationPayload("ORD-001", "TRK123456", "delivered", "Customer Address"),
	}

	msgChan := t.subscribeToNotificationWebSocket(ctx, userID)

	callNotificationForwardingAPI(
		t.T(),
		t.cnt.HTTPServer.GetRunningPort(),
		"/notifications/shipping-update",
		userID,
		shippingUpdates...,
	)

	doneAssertion := assertNotifications(
		t.T(),
		generalnotifications.ShippingUpdateNotification,
		msgChan,
		shippingUpdates...,
	)
	<-doneAssertion
}
