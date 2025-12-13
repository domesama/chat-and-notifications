package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler/config"
	"github.com/domesama/chat-and-notifications/eventmodel"
	"github.com/domesama/chat-and-notifications/outgoinghttp"
	"github.com/domesama/concurrent"
)

// @@wire-struct@@
type ChatMessageSyncService struct {
	Config config.ChatPersistenceChangeHandlerConfig
}

func (c ChatMessageSyncService) ForwardChatMessageToWebsocketServices(
	ctx context.Context,
	msg eventmodel.ChatMessagePersistenceChangeEvent) error {

	// Concurrently forward chat message to relevant service that holds recipients websockets
	forwardChatToSubscribedWebSocket := concurrent.NewTask(
		func(ctx context.Context) error {
			return c.forwardChatToSubscribedWebSocket(ctx, msg)
		},
	)

	forwardChatToGeneralNotificationWebSocket := concurrent.NewTask(
		func(ctx context.Context) error {
			return c.forwardChatToGeneralNotificationWebSocket(ctx, msg)
		},
	)

	err := concurrent.NewGroup(ctx).Exec(forwardChatToSubscribedWebSocket, forwardChatToGeneralNotificationWebSocket)
	return err.ErrorOrNil()
}

func (c ChatMessageSyncService) forwardChatToSubscribedWebSocket(
	ctx context.Context,
	msg eventmodel.ChatMessagePersistenceChangeEvent) (err error) {
	conf := c.Config.ChatMessageSocketTransferOutgoingConfig

	request := outgoinghttp.BuildBasicRequest(
		http.MethodPost,
		conf.Host+"/chat/forward-to-websocket",
		outgoinghttp.WithAdditionalBody(msg.ChatMessage),
	)

	client := &http.Client{Timeout: conf.Timeout}
	_, statusCode, err := outgoinghttp.CallHTTP[any](ctx, client, request)

	// This returns 206 when some websockets did not receive the message and kafkawrapper should automatically retry this message
	if statusCode == http.StatusPartialContent {
		return errors.New("partial content delivered to chat websocket service")
	}

	return
}

func (c ChatMessageSyncService) forwardChatToGeneralNotificationWebSocket(
	ctx context.Context,
	msg eventmodel.ChatMessagePersistenceChangeEvent) (err error) {
	conf := c.Config.GeneralNotificationOutgoingConfig

	request := outgoinghttp.BuildBasicRequest(
		http.MethodPost,
		conf.Host+"/noti/chat",
		outgoinghttp.WithAdditionalBody(msg.ChatMessage),
	)

	client := &http.Client{Timeout: conf.Timeout}
	_, _, err = outgoinghttp.CallHTTP[any](ctx, client, request)
	return
}
