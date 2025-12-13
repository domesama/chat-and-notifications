//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
)

func StartChatPersistenceContainer() (
	ChatPersistenceContainer, func(), error,
) {
	wire.Build(MainBindingSet)
	return ChatPersistenceContainer{}, func() {}, nil
}
