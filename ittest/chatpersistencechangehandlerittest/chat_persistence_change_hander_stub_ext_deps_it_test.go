package ittest

import (
	"context"
	"net/http"

	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/ittest/stub/chatwebsockethandlerstub"
	"github.com/domesama/chat-and-notifications/ittest/stub/generalnotificationhandlerstub"
	"github.com/domesama/chat-and-notifications/model"
)

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
	targetAllRequests := func(ctx context.Context, message model.ChatMessage) bool {
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
