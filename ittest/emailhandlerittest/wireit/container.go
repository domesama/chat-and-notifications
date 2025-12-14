package wireit

import (
	applicationwire "github.com/domesama/chat-and-notifications/cmd/emailhandler/wire"
	"github.com/domesama/chat-and-notifications/email"
	"github.com/domesama/chat-and-notifications/ittest/ittesthelper"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailHandlerITTestContainer struct {
	applicationwire.Locator
	applicationwire.EmailHandlerContainer

	*mongo.Database
	SimpleEmailSender *ittesthelper.SimpleEmailSender
}

var ITTestBindingSet = wire.NewSet(
	applicationwire.MainBindingSet,

	email.ProvideEmailConfig,
	ittesthelper.SimpleEmailSenderSet,

	wire.Struct(new(applicationwire.Locator), "*"),
	wire.Struct(new(EmailHandlerITTestContainer), "*"),
)
