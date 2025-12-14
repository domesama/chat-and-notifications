package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailInfo struct {
	UserID string `bson:"user_id"`
	Email  string `bson:"email"`
	Name   string `bson:"name"`
}

// @@wire-struct@@
type EmailInfoService struct {
	DB *mongo.Database
}

func (r EmailInfoService) GetReceiverEmail(ctx context.Context, receiverID string) (info EmailInfo, err error) {
	err = r.DB.Collection("user_email").FindOne(ctx, bson.M{"user_id": receiverID}).Decode(&info)
	return
}
