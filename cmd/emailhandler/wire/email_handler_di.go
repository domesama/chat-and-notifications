//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
)

func StartEmailHandlerContainer() (
	EmailHandlerContainer, func(), error,
) {
	wire.Build(MainBindingSet, RealSet)
	return EmailHandlerContainer{}, func() {}, nil
}
