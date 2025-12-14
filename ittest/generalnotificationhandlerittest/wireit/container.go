package wireit

import (
	applicationwire "github.com/domesama/chat-and-notifications/cmd/generalnotificationshandler/wire"
	"github.com/google/wire"
)

type GeneralNotificationHandlerITTestContainer struct {
	applicationwire.Locator
	applicationwire.GeneralNotificationHandlerContainer
}

var ITTestBindingSet = wire.NewSet(
	applicationwire.MainBindingSet,
	wire.Struct(new(applicationwire.Locator), "*"),
	wire.Struct(new(GeneralNotificationHandlerITTestContainer), "*"),
)
