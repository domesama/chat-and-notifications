package ittest

import (
	"context"
	"testing"

	"github.com/domesama/chat-and-notifications/event"
	"github.com/domesama/chat-and-notifications/eventmodel"
	"github.com/domesama/chat-and-notifications/ittest/ittesthelper"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChatPersistenceChangeHandlerITTestSuite struct {
	BaseChatPersistenceChangeHandlerITTestSuite
}

func (t *ChatPersistenceChangeHandlerITTestSuite) SetupSuite() {
	t.BaseChatPersistenceChangeHandlerITTestSuite.SetupSuite()
}

func TestChatPersistenceChangeHandlerITTestSuite(t *testing.T) {
	suite.Run(t, new(ChatPersistenceChangeHandlerITTestSuite))
}

func (t *ChatPersistenceChangeHandlerITTestSuite) TestConsumingChange() {

	ctx := context.Background()
	t.metricHelper.ResetEventMetric()

	chatMessages := stub.CreateChatMessages(
		"sender-a", " sender-b",
		"Hello",
		"Should partial on this message", // This message will simulate partial success in WebSocket forwarding
		"Should error on this message",   // This message will simulate error in WebSocket forwarding
		"Goodbye",
	)

	// Test event with duplicated messages to ensure idempotency by event store
	chatMessages = append(chatMessages, addDuplicatedMessage()...)

	// Stub external dependencies: WebSocket forwarding and notification handling to simulate various outcomes based on message content.
	t.stubExternalDependencies()
	t.publishChatPersistenceChangeEvents(ctx, eventmodel.EventTypeCreate, chatMessages...)

	// Assert event processing metrics using EventuallyAssertSelectedCounterMetrics:
	// This method polls Prometheus metrics with retries (up to 5s) until the total count reaches the expected value.
	// We use this instead of direct assertions because events are processed asynchronously via Kafka consumers.
	//
	// Expected outcomes:
	// - 3 success events: "Hello", "Goodbye", and first "Duplicated message"
	// - 2 failed events: "Should partial..." and "Should error..." (non-2xx responses from WebSocket)
	// - 1 dropped event: second "Duplicated message" (event store validation prevents duplicate processing)
	//
	// The final parameter (6) represents the total expected metric count across all categories.
	// This ensures all 6 published events were accounted for before asserting individual metric values.
	t.metricHelper.EventuallyAssertSelectedCounterMetrics(
		map[string][]ittesthelper.Label{
			event.SuccessEventMetricType.GetMetricTotalName(): {
				{LabelName: "event_type", LabelValue: "c", ExpectedValue: 3},
			},
			event.FailedEventMetricType.GetMetricTotalName(): {
				{LabelName: "event_type", LabelValue: "c", ExpectedValue: 2},
			},
			event.DroppedEventMetricType.GetMetricTotalName(): {
				{
					LabelName:     "dropped_reason",
					LabelValue:    event.EventReasonDroppedFromEventStoreValidation.ToString(),
					ExpectedValue: 1,
				},
			},
		}, 6,
	)
}

func addDuplicatedMessage() []model.ChatMessage {
	duplicatedMessage := stub.CreateChatMessages(
		"sender-a", " sender-b", "Duplicated message from kafka",
	)
	duplicatedMessage[0].MessageID = "duplicate-msg-id"

	return append(duplicatedMessage, duplicatedMessage...)
}

func (t *ChatPersistenceChangeHandlerITTestSuite) publishChatPersistenceChangeEvents(
	ctx context.Context,
	eventType eventmodel.ChangeEventType,
	chatMsgs ...model.ChatMessage) {

	rawPayloads := stub.CreateRawChatMongoChangesPayloads(
		t.T(),
		eventType,
		chatMsgs...,
	)

	for i, rawPayload := range rawPayloads {
		rawPayloadBytes, err := json.Marshal(rawPayload)
		assert.NoError(t.T(), err)

		t.kafkaProducer.PublishRawAtMostOnce(
			ctx,
			t.appConfig.KafkaInfo.TopicName,
			chatMsgs[i].MessageID, // Use MessageID as key. Debezium automatically uses there _id as key
			rawPayloadBytes,
			nil,
		)
	}
}
