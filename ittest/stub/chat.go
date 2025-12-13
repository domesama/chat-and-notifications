package stub

import (
	"context"
	"slices"
	"testing"

	"github.com/domesama/chat-and-notifications/chat"
	"github.com/domesama/chat-and-notifications/eventmodel"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateChatMessages(
	sender string,
	receiver string,
	msgContent ...string) []chat.ChatMessage {
	res := make([]chat.ChatMessage, len(msgContent))
	streamID := chat.ComputeStreamID(sender, receiver)
	for i, currentMsgContent := range msgContent {

		chatMessage := chat.ChatMessage{
			Content: currentMsgContent,
			ChatMetadata: chat.ChatMetadata{
				StreamID:   streamID,
				SenderID:   sender,
				ReceiverID: receiver,
			},
		}

		res[i] = chatMessage
	}

	return res
}

func CreateRawChatMongoChangesPayloads(
	t *testing.T,
	eventType eventmodel.ChangeEventType,
	messages ...chat.ChatMessage) []eventmodel.RawMongoChangePayload {
	res := make([]eventmodel.RawMongoChangePayload, len(messages))

	for i, msg := range messages {

		chatBytes, err := bson.MarshalExtJSON(msg, true, true)
		assert.NoError(t, err)

		res[i] = eventmodel.RawMongoChangePayload{
			EventType: eventType,
			After:     ptr.Of(string(chatBytes)),
		}
	}

	return res
}

func WithContainingChatContent(content ...string) Predicate[chat.ChatMessage] {
	return func(ctx context.Context, message chat.ChatMessage) bool {
		return slices.Contains(content, message.Content)
	}
}
