package ittest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/domesama/chat-and-notifications/generalnotifications"
	"github.com/domesama/chat-and-notifications/ittest/generalnotificationhandlerittest/wireit"
	"github.com/domesama/chat-and-notifications/outgoinghttp"
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BaseGeneralNotificationHandlerITTestSuite struct {
	suite.Suite
	cnt wireit.GeneralNotificationHandlerITTestContainer
}

func (t *BaseGeneralNotificationHandlerITTestSuite) SetupSuite() {
	t.NoError(godotenv.Load("../../.env.integration"))

	cnt, cleanUp, err := wireit.InitGeneralNotificationHandlerITTestContainer()
	t.NoError(err)
	t.T().Cleanup(cleanUp)

	t.cnt = cnt
}

// assertNotifications is a generic helper that receives notifications from a channel,
func assertNotifications[T any](
	t *testing.T,
	expectedNotificationType generalnotifications.NotificationType,
	msgChan <-chan generalnotifications.NotificationEnvelope,
	expectedPayloads ...T,
) (done chan bool) {
	actualEnvelopes := make([]generalnotifications.NotificationEnvelope, 0, len(expectedPayloads))
	done = make(chan bool)

	go func() {
		for envelope := range msgChan {
			actualEnvelopes = append(actualEnvelopes, envelope)
			if len(actualEnvelopes) == len(expectedPayloads) {
				break
			}
		}

		for i, envelope := range actualEnvelopes {
			assert.Equal(t, expectedNotificationType, envelope.NotificationType)

			// Unmarshal payload back to struct for proper comparison
			payloadBytes, err := json.Marshal(envelope.Payload)
			assert.NoError(t, err)
			expectedBytes, err := json.Marshal(expectedPayloads[i])
			assert.NoError(t, err)

			assert.JSONEq(t, string(expectedBytes), string(payloadBytes))
		}

		done <- true
	}()

	return done
}

// callNotificationForwardingAPI is a generic helper that forwards notifications to the API
func callNotificationForwardingAPI[T any](
	t *testing.T,
	port string,
	route string,
	userID string,
	payloads ...T,
) {
	ctx := context.Background()
	for _, payload := range payloads {
		req := outgoinghttp.BuildBasicRequest(
			http.MethodPost,
			fmt.Sprintf("http://localhost%s%s", port, route),
			outgoinghttp.WithAdditionalBody(payload),
			outgoinghttp.WithAdditionalQuery(url.Values{"user_id": {userID}}),
		)

		client := &http.Client{}
		_, statusCode, err := outgoinghttp.CallHTTP[map[string]interface{}](ctx, client, req)

		assert.Equal(t, http.StatusOK, statusCode)
		assert.NoError(t, err)
	}
}

// subscribeToNotificationWebSocket is a generic helper that subscribes to a WebSocket endpoint
func (t *BaseGeneralNotificationHandlerITTestSuite) subscribeToNotificationWebSocket(
	ctx context.Context,
	userID string,
) chan generalnotifications.NotificationEnvelope {
	port := t.cnt.HTTPServer.GetRunningPort()

	wsURL := fmt.Sprintf("ws://localhost%s/notifications/subscribe?user_id=%s", port, userID)

	msgChan, _, err := websocket.SubscribeToWebSocket[generalnotifications.NotificationEnvelope](ctx, wsURL)
	t.NoError(err)

	return msgChan
}
