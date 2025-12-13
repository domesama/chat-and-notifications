package ittest

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/domesama/chat-and-notifications/chat"
	"github.com/domesama/chat-and-notifications/outgoinghttp"
)

func (t *ChatPersistenceITTestSuite) TestInvalidMessages() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		message     chat.ChatMessage
		description string
	}{
		{
			name: "EmptyContent",
			message: chat.ChatMessage{
				Content: "",
				ChatMetadata: chat.ChatMetadata{
					StreamID:   chat.ComputeStreamID("sender-1", "receiver-1"),
					SenderID:   "sender-1",
					ReceiverID: "receiver-1",
				},
			},
			description: "empty message content should be rejected",
		},
		{
			name: "MissingStreamID",
			message: chat.ChatMessage{
				Content: "Hello",
				ChatMetadata: chat.ChatMetadata{
					StreamID:   "",
					SenderID:   "sender-1",
					ReceiverID: "receiver-1",
				},
			},
			description: "missing stream_id should be rejected",
		},
		{
			name: "MissingSenderID",
			message: chat.ChatMessage{
				Content: "Hello",
				ChatMetadata: chat.ChatMetadata{
					StreamID:   chat.ComputeStreamID("sender-1", "receiver-1"),
					SenderID:   "",
					ReceiverID: "receiver-1",
				},
			},
			description: "missing sender_id should be rejected",
		},
		{
			name: "MissingReceiverID",
			message: chat.ChatMessage{
				Content: "Hello",
				ChatMetadata: chat.ChatMetadata{
					StreamID:   chat.ComputeStreamID("sender-1", "receiver-1"),
					SenderID:   "sender-1",
					ReceiverID: "",
				},
			},
			description: "missing receiver_id should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func() {
				_, statusCode := t.callChatPersistenceAPI(ctx, tc.message)
				t.Equal(http.StatusBadRequest, statusCode, tc.description)
			},
		)
	}
}

// TestValidContentVariations tests that various valid content types are handled correctly
func (t *ChatPersistenceITTestSuite) TestValidContentVariations() {
	ctx := context.Background()

	testCases := []struct {
		name     string
		content  string
		senderID string
	}{
		{
			name:     "SpecialCharacters",
			content:  "Hello! @#$%^&*()_+-={}[]|\\:\";<>?,./~`",
			senderID: "sender-2",
		},
		{
			name:     "UnicodeContent",
			content:  "สวัสดี 你好 こんにちは 안녕하세요 🎉🚀💻",
			senderID: "sender-3",
		},
		{
			name:     "VeryLongContent",
			content:  strings.Repeat("This is a very long message. ", 350),
			senderID: "sender-4",
		},
		{
			name:     "WhitespaceOnly",
			content:  "   \t\n   ",
			senderID: "sender-6",
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func() {
				receiverID := "receiver-" + tc.senderID[7:] // Extract number from sender ID
				chatMessage := chat.ChatMessage{
					Content: tc.content,
					ChatMetadata: chat.ChatMetadata{
						StreamID:   chat.ComputeStreamID(tc.senderID, receiverID),
						SenderID:   tc.senderID,
						ReceiverID: receiverID,
					},
				}

				resp, statusCode := t.callChatPersistenceAPI(ctx, chatMessage)
				t.Equal(http.StatusCreated, statusCode)
				t.Equal(chatMessage.ChatMetadata, resp.ChatMetadata)

				// Verify the message was stored correctly
				t.assertChatMessageInDatabase(ctx, chatMessage.StreamID, chatMessage)
			},
		)
	}
}

func (t *ChatPersistenceITTestSuite) callChatPersistenceAPI(ctx context.Context,
	message chat.ChatMessage) (chat.ChatMessage, int) {
	port := t.cnt.HTTPServer.GetRunningPort()

	req := outgoinghttp.BuildBasicRequest(
		http.MethodPost,
		fmt.Sprintf("http://localhost%s/chat/persist", port),
		outgoinghttp.WithAdditionalBody(message),
	)

	client := &http.Client{}
	resp, statusCode, _ := outgoinghttp.CallHTTP[chat.ChatMessage](ctx, client, req)

	return resp, statusCode
}
