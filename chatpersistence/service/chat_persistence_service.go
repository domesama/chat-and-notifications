package service

import (
	"context"
	"time"

	"github.com/domesama/chat-and-notifications/model"
	"go.mongodb.org/mongo-driver/mongo"
)

// @@wire-struct@@
type ChatPersistenceService struct {
	DB *mongo.Database
}

func (c ChatPersistenceService) PersistChatMessage(ctx context.Context, message model.ChatMessage) (err error) {
	message.CreatedAt = time.Now()
	_, err = c.DB.Collection("chat").InsertOne(ctx, message)
	return
}
