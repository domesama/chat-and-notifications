package ittest

import (
	"fmt"
	"os"

	"github.com/domesama/chat-and-notifications/ittest/chatpersistenceittest/wireit"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type BaseChatPersistenceITTestSuite struct {
	suite.Suite
	cnt wireit.ChatPersistenceITTestContainer
}

func (t *BaseChatPersistenceITTestSuite) SetupSuite() {
	t.NoError(godotenv.Load("../../.env.integration"))

	t.WithPrefixedMongoDatabase()
	cnt, cleanUp, err := wireit.InitChatPersistenceITTestContainer()
	t.NoError(err)
	t.T().Cleanup(cleanUp)

	t.cnt = cnt
}

func (t *BaseChatPersistenceITTestSuite) WithPrefixedMongoDatabase() {
	t.NoError(os.Setenv("MONGO_DATABASE", fmt.Sprintf("%s_%d", t.T().Name(), os.Getpid())))
}
