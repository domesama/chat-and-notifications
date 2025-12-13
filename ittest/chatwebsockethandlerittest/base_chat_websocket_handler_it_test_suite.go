package ittest

import (
	"github.com/domesama/chat-and-notifications/ittest/chatwebsockethandlerittest/wireit"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type BaseChatWebSocketHandlerITTestSuite struct {
	suite.Suite
	cnt wireit.ChatWebSocketHandlerITTestContainer
}

func (t *BaseChatWebSocketHandlerITTestSuite) SetupSuite() {
	t.NoError(godotenv.Load("../../.env.integration"))

	cnt, cleanUp, err := wireit.InitChatWebSocketHandlerITTestContainer()
	t.NoError(err)
	t.T().Cleanup(cleanUp)

	t.cnt = cnt
}
