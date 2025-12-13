package chatpersistencechangehandler

import (
	"context"

	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler/service"
	"github.com/domesama/chat-and-notifications/event/eventmsg"
	"github.com/domesama/chat-and-notifications/eventmodel"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/bson"
)

// @@wire-struct@@
type ChatPersistenceChangeMessageHandler struct {
	ChatMessageSyncService service.ChatMessageSyncService
}

func (c ChatPersistenceChangeMessageHandler) GetEventType(msg eventmsg.Message[eventmodel.ChatMessagePersistenceChangeEvent]) string {
	return string(msg.Value.EventType)
}

func (c ChatPersistenceChangeMessageHandler) ConvertMessageValue(
	rawValues []byte,
) (
	eventValue eventmodel.ChatMessagePersistenceChangeEvent, shouldDrop bool, err error,
) {

	var rawMongoChange eventmodel.RawMongoChangePayload
	err = json.Unmarshal(rawValues, &rawMongoChange)

	if err != nil || rawMongoChange.After == nil || rawMongoChange.EventType == "" {
		shouldDrop = true
		return
	}

	eventValue.EventType = rawMongoChange.EventType

	err = bson.UnmarshalExtJSON([]byte(*rawMongoChange.After), true, &eventValue.ChatMessage)
	if err != nil {
		shouldDrop = true
	}

	return
}

func (c ChatPersistenceChangeMessageHandler) HandleMessage(
	ctx context.Context,
	msg eventmsg.Message[eventmodel.ChatMessagePersistenceChangeEvent],
) (err error) {

	switch msg.Value.EventType {
	case eventmodel.EventTypeCreate:
		return c.ChatMessageSyncService.ForwardChatMessageToWebsocketServices(ctx, msg.Value)
	default:
		// No-op for other event types yet, we can support chat message delete and update here accordingly.
		return
	}
}
