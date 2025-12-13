package connectionconfig

import "github.com/domesama/kafkawrapper"

type KafkaConsumerInfo struct {
	TopicName          string                   `envconfig:"TOPIC_NAME" required:"true"`
	ConsumerName       string                   `envconfig:"CONSUMER_NAME" required:"true"`
	IgnoreOldMessage   bool                     `envconfig:"IGNORE_OLD_MESSAGE" default:"false"`
	MessageRetryConfig kafkawrapper.RetryConfig `envconfig:"MESSAGE_RETRY"`
}
