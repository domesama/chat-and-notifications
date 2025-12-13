package handler

import (
	"github.com/domesama/chat-and-notifications/httpserverwrapper"
	"github.com/gin-gonic/gin"
)

func ProvideRouterCustomizer(handler ChatPersistenceHandler) httpserverwrapper.RouterCustomizer {
	return handler
}

func (c ChatPersistenceHandler) Configure(builder *httpserverwrapper.HTTPServerBuilder) error {
	builder.WithMiddleware(gin.Recovery())
	return nil
}

func (c ChatPersistenceHandler) RegisterRoutes(engine *gin.Engine) error {
	chatRouterGroup := engine.Group("/chat")
	{
		chatRouterGroup.POST("/persist", c.HandleChatPersistence)
	}
	return nil
}
