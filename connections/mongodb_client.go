package connections

import (
	"context"
	"fmt"
	"time"

	"github.com/domesama/chat-and-notifications/connections/connectionconfig"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoSet = wire.NewSet(
	connectionconfig.ProvideMongoDBConfig,

	ProvideMongoClient,
	ProvideMongoDatabase,
)

func ProvideMongoClient(cfg connectionconfig.MongoDBConfig) (*mongo.Client, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(cfg.URI)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, func() {}, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}

	return client, cleanup, nil
}

func ProvideMongoDatabase(client *mongo.Client, cfg connectionconfig.MongoDBConfig) *mongo.Database {
	return client.Database(cfg.Database)
}
