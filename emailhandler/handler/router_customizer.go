package handler

import (
	"github.com/domesama/chat-and-notifications/httpserverwrapper"
	"github.com/gin-gonic/gin"
)

func ProvideRouterCustomizer(handler EmailHandler) httpserverwrapper.RouterCustomizer {
	return handler
}

func (c EmailHandler) Configure(builder *httpserverwrapper.HTTPServerBuilder) error {
	builder.WithMiddleware(gin.Recovery())
	return nil
}

func (c EmailHandler) RegisterRoutes(engine *gin.Engine) error {
	chatRouterGroup := engine.Group("/email")
	{
		chatRouterGroup.POST("/chat", c.HandleChatMailing)
		chatRouterGroup.POST("/purchased", c.HandlePurchaseMailing)
	}
	return nil
}
