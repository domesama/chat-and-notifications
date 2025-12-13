package ittest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/domesama/chat-and-notifications/chat"
	"github.com/domesama/chat-and-notifications/event"
	"github.com/domesama/chat-and-notifications/eventmodel"
	"github.com/domesama/chat-and-notifications/ittest/ittesthelper"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/ittest/stub/chatwebsockethandlerstub"
	"github.com/domesama/chat-and-notifications/ittest/stub/generalnotificationhandlerstub"
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
		"Should partial on this message",
		"Should error on this message",
		"Goodbye",
	)

	rawPayloads := stub.CreateRawChatMongoChangesPayloads(
		t.T(),
		eventmodel.EventTypeCreate,
		chatMessages...,
	)

	t.stubExternalDependencies()
	t.publishChatPersistenceChangeEvents(ctx, rawPayloads...)

	t.metricHelper.EventuallyAssertSelectedCounterMetrics(
		map[string][]ittesthelper.Label{
			event.SuccessEventMetricType.GetMetricTotalName(): {
				{LabelName: "event_type", LabelValue: "c", ExpectedValue: 2},
			},
			event.FailedEventMetricType.GetMetricTotalName(): {
				{LabelName: "event_type", LabelValue: "c", ExpectedValue: 2},
			},
		}, 4,
	)
}

func (t *ChatPersistenceChangeHandlerITTestSuite) stubExternalDependencies() {
	t.stubForwardChatMessageToWebSocket()
	t.stubChatNotification()
}

func (t *ChatPersistenceChangeHandlerITTestSuite) stubForwardChatMessageToWebSocket() {
	validMsgContents := []string{"Hello", "Goodbye"}
	partialWebSocketMsgContent := "Should partial on this message"
	errorWebSocketMsgContent := "Should error on this message"

	shouldReturn200WhenSeeingTheseMsgContents := chatwebsockethandlerstub.ChatWebSocketForwardAPIStub{
		Predicates: stub.NewPredicates(
			stub.WithContainingChatContent(validMsgContents...),
		),
		StubStatusCode: http.StatusOK,
	}

	shouldReturn500WhenSeeingTheseMsgContents := chatwebsockethandlerstub.ChatWebSocketForwardAPIStub{
		Predicates: stub.NewPredicates(
			stub.WithContainingChatContent(partialWebSocketMsgContent),
		),
		StubStatusCode: http.StatusPartialContent,
	}

	shouldReturn206WhenSeeingTheseMsgContents := chatwebsockethandlerstub.ChatWebSocketForwardAPIStub{
		Predicates: stub.NewPredicates(
			stub.WithContainingChatContent(errorWebSocketMsgContent),
		),
		StubStatusCode: http.StatusInternalServerError,
	}

	t.externalDependencies.ChatWebSocketForwarderStub.AddForwardChatMessageToWebSocketStub(
		shouldReturn200WhenSeeingTheseMsgContents,
		shouldReturn206WhenSeeingTheseMsgContents,
		shouldReturn500WhenSeeingTheseMsgContents,
	)
}

func (t *ChatPersistenceChangeHandlerITTestSuite) stubChatNotification() {
	targetAllRequests := func(ctx context.Context, message chat.ChatMessage) bool {
		return true
	}

	alwaysReturn200 := generalnotificationhandlerstub.ChatNotificationStub{
		Predicates:     stub.NewPredicates(targetAllRequests),
		StubStatusCode: http.StatusOK,
	}

	t.externalDependencies.ChatNotificationStub.AddForwardChatMessageToWebSocketStub(
		alwaysReturn200,
	)

}

func (t *ChatPersistenceChangeHandlerITTestSuite) publishChatPersistenceChangeEvents(
	ctx context.Context,
	rawPayloads ...eventmodel.RawMongoChangePayload) {

	for i, rawPayload := range rawPayloads {
		rawPayloadBytes, err := json.Marshal(rawPayload)
		assert.NoError(t.T(), err)

		t.kafkaProducer.PublishRawAtMostOnce(
			ctx,
			t.appConfig.KafkaInfo.TopicName,
			fmt.Sprint(i), rawPayloadBytes,
			nil,
		)
	}
}
