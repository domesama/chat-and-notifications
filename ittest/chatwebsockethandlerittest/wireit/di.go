//go:build wireinject
// +build wireinject

package wireit

import (
	"github.com/google/wire"
)

func InitChatWebSocketHandlerITTestContainer() (
	ChatWebSocketHandlerITTestContainer, func(), error,
) {
	wire.Build(ITTestBindingSet)
	return ChatWebSocketHandlerITTestContainer{}, func() {}, nil
}
