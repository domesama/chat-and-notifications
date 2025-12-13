package chatpersistencechangehandler

import "github.com/domesama/chat-and-notifications/event"

type (
	ChatPersistenceChangeEventMetric *event.EventMetric
)

func ProvideChatPersistenceChangeEventMetric() ChatPersistenceChangeEventMetric {
	return event.CreateEventMetrics("chat_persistence_change_consumer")
}
