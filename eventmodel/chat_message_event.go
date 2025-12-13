package eventmodel

import "github.com/domesama/chat-and-notifications/chat"

type ChatMessagePersistenceChangeEvent struct {
	EventType   ChangeEventType
	ChatMessage chat.ChatMessage
}
