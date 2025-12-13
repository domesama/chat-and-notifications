package ittest

import (
	"strconv"

	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler/config"
	"github.com/domesama/chat-and-notifications/ittest/ittesthelper"
	"github.com/domesama/chat-and-notifications/ittest/stub/chatwebsockethandlerstub"
	"github.com/domesama/chat-and-notifications/ittest/stub/generalnotificationhandlerstub"
	"github.com/domesama/chat-and-notifications/outgoinghttp"
	"github.com/gin-gonic/gin"
)

type ExternalDependencies struct {
	ChatWebSocketForwarderStub *chatwebsockethandlerstub.ChatWebSocketForwarderStub
	ChatNotificationStub       *generalnotificationhandlerstub.ChatNotificationHandlerStub
}

func (t *BaseChatPersistenceChangeHandlerITTestSuite) StartHTTPStubServers(
	appConf *config.ChatPersistenceChangeHandlerConfig) (
	extDeps ExternalDependencies,
) {

	startedPort := ittesthelper.StartHTTPServer(
		t.T(), func(engine *gin.Engine) error {

			extDeps.ChatWebSocketForwarderStub = chatwebsockethandlerstub.ChatWebSocketHandlerRouterStub(engine)
			extDeps.ChatNotificationStub = generalnotificationhandlerstub.GeneralNotificationRouterStub(engine)

			return nil
		},
	)

	startedHost := "http://localhost:" + strconv.Itoa(startedPort)

	appConf.ChatMessageSocketTransferOutgoingConfig = outgoinghttp.OutGoingHTTPConfig{Host: startedHost}
	appConf.GeneralNotificationOutgoingConfig = outgoinghttp.OutGoingHTTPConfig{Host: startedHost}

	return extDeps
}
