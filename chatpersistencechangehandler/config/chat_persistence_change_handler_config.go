package config

import (
	"github.com/domesama/chat-and-notifications/connections/connectionconfig"
	"github.com/domesama/chat-and-notifications/outgoinghttp"
	"github.com/domesama/kafkawrapper"
)

type ChatPersistenceChangeHandlerConfig struct {
	KafkaInfo             connectionconfig.KafkaConsumerInfo `envconfig:"CHAT_PERSISTENCE_CHANGE_KAFKA_CONSUMER_INFO"`
	KafkaConnectionConfig kafkawrapper.KafkaConfig           `envconfig:"CHAT_PERSISTENCE_CHANGE"`

	GeneralNotificationOutgoingConfig       outgoinghttp.OutGoingHTTPConfig `envconfig:"GENERAL_NOTIFICATION_OUTGOING_CONFIG" required:"true"`
	ChatMessageSocketTransferOutgoingConfig outgoinghttp.OutGoingHTTPConfig `envconfig:"CHAT_MESSAGE_SOCKET_TRANSFER_OUTGOING_CONFIG" required:"true"`
}
