//go:build wireinject
// +build wireinject

package wireit

import (
	"github.com/domesama/chat-and-notifications/chatpersistencechangehandler/config"
	"github.com/google/wire"
)

func InitChatPersistenceChangeHandlerITTestContainer(_ config.ChatPersistenceChangeHandlerConfig) (
	ChatPersistenceChangeHandlerITTestContainer, func(), error,
) {
	wire.Build(ITTestBindingSet)
	return ChatPersistenceChangeHandlerITTestContainer{}, func() {}, nil
}
