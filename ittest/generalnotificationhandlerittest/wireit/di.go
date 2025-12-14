//go:build wireinject
// +build wireinject

package wireit

import (
	"github.com/google/wire"
)

func InitGeneralNotificationHandlerITTestContainer() (
	GeneralNotificationHandlerITTestContainer, func(), error,
) {
	wire.Build(ITTestBindingSet)
	return GeneralNotificationHandlerITTestContainer{}, func() {}, nil
}
