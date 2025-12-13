package service

import (
	"context"

	"github.com/domesama/chat-and-notifications/chat"
	"go.mongodb.org/mongo-driver/mongo"
)

// @@wire-struct@@
type ChatPersistenceService struct {
	DB *mongo.Database
}

func (c ChatPersistenceService) PersistChatMessage(ctx context.Context, message chat.ChatMessage) (err error) {
	_, err = c.DB.Collection("chat").InsertOne(ctx, message)
	return
}
