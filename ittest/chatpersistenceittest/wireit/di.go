//go:build wireinject
// +build wireinject

package wireit

import (
	"github.com/google/wire"
)

func InitChatPersistenceITTestContainer() (
	ChatPersistenceITTestContainer, func(), error,
) {
	wire.Build(ITTestBindingSet)
	return ChatPersistenceITTestContainer{}, func() {}, nil
}
