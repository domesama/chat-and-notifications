package ittest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/domesama/chat-and-notifications/chat"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/outgoinghttp"
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/stretchr/testify/suite"
)

type ChatWebSocketHandlerITTestSuite struct {
	BaseChatWebSocketHandlerITTestSuite
}

func (t *ChatWebSocketHandlerITTestSuite) SetupSuite() {
	t.BaseChatWebSocketHandlerITTestSuite.SetupSuite()
}

func TestChatWebSocketHandlerITTestSuite(t *testing.T) {
	suite.Run(t, new(ChatWebSocketHandlerITTestSuite))
}

func (t *ChatWebSocketHandlerITTestSuite) TestChatWebSocketCommunications() {
	ctx := context.Background()

	chatMessages := stub.CreateChatMessages(
		"Mr.A", "Mr.B",
		"Hello",
		"How are you?",
		"Goodbye",
	)

	// Subscribe to WebSocket before sending messages
	msgChanAToB := t.subscribeToChatWebSocket(ctx, "Mr.A", "Mr.B")
	msgChanBToA := t.subscribeToChatWebSocket(ctx, "Mr.B", "Mr.A")

	// In case Mr.A is connected from multiple devices
	msgChanAOtherDeviceToB := t.subscribeToChatWebSocket(ctx, "Mr.A", "Mr.B")

	// Forward messages to WebSocket subscribers
	t.callChatSocketForwardingAPI(ctx, chatMessages...)

	doneAssertionAToB := t.assertChatMessages(msgChanAToB, chatMessages...)
	doneAssertionBToA := t.assertChatMessages(msgChanBToA, chatMessages...)
	doneAssertionAOtherDeviceToB := t.assertChatMessages(msgChanAOtherDeviceToB, chatMessages...)

	<-doneAssertionAToB
	<-doneAssertionBToA
	<-doneAssertionAOtherDeviceToB
}

func (t *ChatWebSocketHandlerITTestSuite) subscribeToChatWebSocket(
	ctx context.Context,
	senderID string,
	receiverID string,
) chan chat.ChatMessage {
	port := t.cnt.HTTPServer.GetRunningPort()

	metadata := chat.ChatMetadata{
		StreamID:   chat.ComputeStreamID(senderID, receiverID),
		SenderID:   senderID,
		ReceiverID: receiverID,
	}

	wsURL := fmt.Sprintf(
		"ws://localhost%s/chat/subscribe-websocket?stream_id=%s&sender_id=%s&receiver_id=%s",
		port, metadata.StreamID, metadata.SenderID, metadata.ReceiverID,
	)

	msgChan, cleanup, err := websocket.SubscribeToWebSocket[chat.ChatMessage](ctx, wsURL)
	t.NoError(err)
	t.T().Cleanup(cleanup)

	return msgChan
}

func (t *ChatWebSocketHandlerITTestSuite) callChatSocketForwardingAPI(ctx context.Context,
	message ...chat.ChatMessage) {
	port := t.cnt.HTTPServer.GetRunningPort()

	for _, msg := range message {
		req := outgoinghttp.BuildBasicRequest(
			http.MethodPost,
			fmt.Sprintf("http://localhost%s/chat/forward-to-websocket", port),
			outgoinghttp.WithAdditionalBody(msg),
		)

		client := &http.Client{}
		_, statusCode, err := outgoinghttp.CallHTTP[map[string]interface{}](ctx, client, req)

		t.Equal(http.StatusOK, statusCode)
		t.NoError(err)
	}

}

func (t *ChatWebSocketHandlerITTestSuite) assertChatMessages(
	msgChan <-chan chat.ChatMessage,
	expectedMessages ...chat.ChatMessage) (done chan bool) {
	actualMessages := make([]chat.ChatMessage, 0, len(expectedMessages))
	done = make(chan bool)

	go func() {
		for msg := range msgChan {
			actualMessages = append(actualMessages, msg)
			if len(actualMessages) == len(expectedMessages) {
				break
			}
		}
		t.Equal(expectedMessages, actualMessages)
		done <- true
	}()

	return done
}
