package ittest

import (
	"context"
	"time"

	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler/config"
	"github.com/domesama/chat-and-notifications/ittest/chatpersistencechangehandlerittest/wireit"
	"github.com/domesama/chat-and-notifications/ittest/ittesthelper"
	"github.com/domesama/kafkawrapper"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/suite"
)

type BaseChatPersistenceChangeHandlerITTestSuite struct {
	suite.Suite
	cnt       wireit.ChatPersistenceChangeHandlerITTestContainer
	appConfig config.ChatPersistenceChangeHandlerConfig

	externalDependencies ExternalDependencies
	kafkaProducer        kafkawrapper.AsyncProducer
	metricHelper         ittesthelper.EventMetricHelper
}

func (t *BaseChatPersistenceChangeHandlerITTestSuite) SetupSuite() {
	t.NoError(godotenv.Load("../../.env.integration"))

	var appConf config.ChatPersistenceChangeHandlerConfig
	envconfig.MustProcess("", &appConf)

	t.WithKafkaConsumerInfo(&appConf)
	t.externalDependencies = t.StartHTTPStubServers(&appConf)

	cnt, cleanUp, err := wireit.InitChatPersistenceChangeHandlerITTestContainer(appConf)
	t.NoError(err)
	t.T().Cleanup(cleanUp)

	t.cnt = cnt
	t.appConfig = appConf

	t.metricHelper = ittesthelper.NewMetricHelper(
		t.T(),
		cnt.GetMonitoringServer(),
		cnt.ChatPersistenceChangeEventMetric,
	)

	producer, err := kafkawrapper.NewAsyncProducer(appConf.KafkaConnectionConfig)
	t.NoError(err)
	t.kafkaProducer = producer

	// Wait till consumers are ready
	t.Eventually(
		func() bool {
			return kafkawrapper.ConsumerGroup(t.cnt.ChatPersistenceChangeHandlerContainer.ChatPersistenceChangeHandler).IsRunning()
		}, 5*time.Second, 500*time.Millisecond, "Consumer group is not ready",
	)

	// Flush all keys in the event store Redis before each suite
	t.cnt.RedisClient.FlushAll(context.Background())
}

func (t *BaseChatPersistenceChangeHandlerITTestSuite) WithKafkaConsumerInfo(cfg *config.ChatPersistenceChangeHandlerConfig) {
	cfg.KafkaInfo = ittesthelper.SuffixKafkaTopicName(t.T().Name())
}
