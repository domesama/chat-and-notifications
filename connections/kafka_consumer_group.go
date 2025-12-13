package connections

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/domesama/chat-and-notifications/connections/connectionconfig"
	doakes "github.com/domesama/doakes/server"
	"github.com/domesama/kafkawrapper"
)

func NewConsumerGroup(
	kafkaCfg kafkawrapper.KafkaConfig,
	kafkaInfo connectionconfig.KafkaConsumerInfo,
	monitoringServer *doakes.TelemetryServer,
	handler kafkawrapper.MessageHandler[*sarama.ConsumerMessage],
) (kafkawrapper.ConsumerGroup, func(), error) {
	consumerName := kafkaInfo.ConsumerName
	topicName := kafkaInfo.TopicName
	ignoreOldMessage := kafkaInfo.IgnoreOldMessage

	wrappedHandler := kafkawrapper.WrapWithRetryBackoffHandler(handler, kafkaInfo.MessageRetryConfig)

	consumer, err := kafkawrapper.NewConsumerGroup(kafkaCfg, consumerName, topicName, ignoreOldMessage, wrappedHandler)
	if err != nil {
		return nil, func() {}, err
	}

	monitoringServer.RegisterHealthCheck(
		consumerName, func() error {
			if consumer.IsRunning() {
				return nil
			}
			return fmt.Errorf("failed to connect kafka consumer %v", consumerName)
		},
	)

	return consumer, func() { consumer.Close() }, consumer.Start()
}
