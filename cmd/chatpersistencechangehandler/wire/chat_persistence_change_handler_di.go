//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler/config"
	"github.com/google/wire"
)

func StartChatPersistenceChangeHandlerContainer(_ config.ChatPersistenceChangeHandlerConfig) (
	ChatPersistenceChangeHandlerContainer, func(), error,
) {
	wire.Build(MainBindingSet)
	return ChatPersistenceChangeHandlerContainer{}, func() {}, nil
}
