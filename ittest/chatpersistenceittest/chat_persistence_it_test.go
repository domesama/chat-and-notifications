package ittest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/domesama/chat-and-notifications/chat"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/outgoinghttp"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatPersistenceITTestSuite struct {
	BaseChatPersistenceITTestSuite
}

func (t *ChatPersistenceITTestSuite) SetupSuite() {
	t.BaseChatPersistenceITTestSuite.SetupSuite()
}

func TestChatPersistenceITTestSuite(t *testing.T) {
	suite.Run(t, new(ChatPersistenceITTestSuite))
}

func (t *ChatPersistenceITTestSuite) TestPersistingDataToMongo() {
	ctx := context.Background()

	chatMessages := stub.CreateChatMessages(
		"sender-a", "receiver-b",
		"Hello",
		"How are you?",
		"Goodbye",
	)

	t.callChatPersistenceServer(ctx, chatMessages...)

	t.assertChatMessageInDatabase(ctx, chat.ComputeStreamID("sender-a", "receiver-b"), chatMessages...)
}

func (t *ChatPersistenceITTestSuite) callChatPersistenceServer(ctx context.Context, message ...chat.ChatMessage) {
	port := t.cnt.HTTPServer.GetRunningPort()

	for _, msg := range message {
		req := outgoinghttp.BuildBasicRequest(
			http.MethodPost,
			fmt.Sprintf("http://localhost%s/chat/persist", port),
			outgoinghttp.WithAdditionalBody(msg),
		)

		client := &http.Client{}
		resp, statusCode, err := outgoinghttp.CallHTTP[chat.ChatMessage](ctx, client, req)

		t.Equal(statusCode, http.StatusCreated)
		t.NoError(err)
		t.Equal(msg.ChatMetadata, resp.ChatMetadata)
	}

}

func (t *ChatPersistenceITTestSuite) assertChatMessageInDatabase(
	ctx context.Context,
	streamID string,
	expectedChatMessages ...chat.ChatMessage,
) {

	var chatMessages []chat.ChatMessage

	cursor, err := t.cnt.Database.Collection("chat").Find(
		ctx,
		bson.M{"stream_id": streamID},
		options.Find().SetSort(bson.M{"created_at": 1}),
	)
	t.NoError(err)
	t.NoError(cursor.All(ctx, &chatMessages))

	for i := range expectedChatMessages {
		t.Equal(expectedChatMessages[i].Content, chatMessages[i].Content)
	}
}
