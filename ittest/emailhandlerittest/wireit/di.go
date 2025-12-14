//go:build wireinject
// +build wireinject

package wireit

import (
	"github.com/google/wire"
)

func InitEmailHandlerITTestContainer() (
	EmailHandlerITTestContainer, func(), error,
) {
	wire.Build(ITTestBindingSet)
	return EmailHandlerITTestContainer{}, func() {}, nil
}
