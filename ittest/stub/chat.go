package stub

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/domesama/chat-and-notifications/chatstream"
	"github.com/domesama/chat-and-notifications/eventmodel"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateChatMessages(
	sender string,
	receiver string,
	msgContent ...string) []model.ChatMessage {
	res := make([]model.ChatMessage, len(msgContent))
	streamID := chatstream.ComputeStreamID(sender, receiver)
	for i, currentMsgContent := range msgContent {

		chatMessage := model.ChatMessage{
			MessageID: fmt.Sprintf("message_id_%d", i),
			Content:   currentMsgContent,
			ChatMetadata: model.ChatMetadata{
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
	messages ...model.ChatMessage) []eventmodel.RawMongoChangePayload {
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

func WithContainingChatContent(content ...string) Predicate[model.ChatMessage] {
	return func(ctx context.Context, message model.ChatMessage) bool {
		return slices.Contains(content, message.Content)
	}
}
