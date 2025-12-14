//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
)

func StartGeneralNotificationHandlerContainer() (
	GeneralNotificationHandlerContainer, func(), error,
) {
	wire.Build(MainBindingSet)
	return GeneralNotificationHandlerContainer{}, func() {}, nil
}
