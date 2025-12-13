//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
)

func StartChatWebSocketHandlerContainer() (
	ChatWebSocketHandlerContainer, func(), error,
) {
	wire.Build(MainBindingSet)
	return ChatWebSocketHandlerContainer{}, func() {}, nil
}
