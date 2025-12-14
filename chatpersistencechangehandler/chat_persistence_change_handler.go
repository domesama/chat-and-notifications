package chatpersistencechangehandler

import (
	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler/config"
	"github.com/domesama/chat-and-notifications/connections"
	"github.com/domesama/chat-and-notifications/event"
	"github.com/domesama/chat-and-notifications/eventmodel"
	doakes "github.com/domesama/doakes/server"
	"github.com/domesama/kafkawrapper"
)

type ChatPersistenceChangeHandler kafkawrapper.ConsumerGroup

func ProvideChatPersistenceChangeHandler(
	conf config.ChatPersistenceChangeHandlerConfig,
	server *doakes.TelemetryServer,
	msgHandler ChatPersistenceChangeMessageHandler,
	metric ChatPersistenceChangeEventMetric,
	eventStore ChatPersistenceChangeEventStore,
) (ChatPersistenceChangeHandler, func(), error) {
	eventHandler := event.NewSingleEventHandler[eventmodel.ChatMessagePersistenceChangeEvent](
		msgHandler,
		metric,
		event.WithEventStore(eventStore),
	)

	return connections.NewConsumerGroup(
		conf.KafkaConnectionConfig,
		conf.KafkaInfo,
		server,
		eventHandler.HandleEvent,
	)
}
