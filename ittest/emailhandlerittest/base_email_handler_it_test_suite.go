package ittest

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/domesama/chat-and-notifications/emailhandler/service"
	"github.com/domesama/chat-and-notifications/ittest/emailhandlerittest/wireit"
	"github.com/domesama/chat-and-notifications/outgoinghttp"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type BaseEmailHandlerITTestSuite struct {
	suite.Suite
	cnt wireit.EmailHandlerITTestContainer
}

func (t *BaseEmailHandlerITTestSuite) SetupSuite() {
	t.NoError(godotenv.Load("../../.env.integration"))

	t.WithPrefixedMongoDatabase()
	cnt, cleanUp, err := wireit.InitEmailHandlerITTestContainer()
	t.NoError(err)
	t.T().Cleanup(cleanUp)

	t.cnt = cnt
}

func (t *BaseEmailHandlerITTestSuite) WithPrefixedMongoDatabase() {
	t.NoError(os.Setenv("MONGO_DATABASE", fmt.Sprintf("%s_%d", t.T().Name(), os.Getpid())))
}

func (t *BaseEmailHandlerITTestSuite) callEmailRoute(
	ctx context.Context,
	route string,
	expectedStatusCode int,
	body interface{},
) {
	port := t.cnt.HTTPServer.GetRunningPort()

	req := outgoinghttp.BuildBasicRequest(
		http.MethodPost,
		fmt.Sprintf("http://localhost%s%s", port, route),
		outgoinghttp.WithAdditionalBody(body),
	)

	client := &http.Client{}
	_, statusCode, _ := outgoinghttp.CallHTTP[map[string]interface{}](ctx, client, req)

	t.Equal(expectedStatusCode, statusCode)
}

func (t *BaseEmailHandlerITTestSuite) insertEmailInfo(ctx context.Context, emailInfo service.EmailInfo) {
	collection := t.cnt.Database.Collection("user_email")
	_, err := collection.InsertOne(ctx, emailInfo)
	t.NoError(err)
}
