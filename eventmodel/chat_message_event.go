package eventmodel

import "github.com/domesama/chat-and-notifications/model"

type ChatMessagePersistenceChangeEvent struct {
	EventType   ChangeEventType
	ChatMessage model.ChatMessage
}
