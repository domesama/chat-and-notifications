package handler

import (
	"github.com/domesama/chat-and-notifications/httpserverwrapper"
	"github.com/gin-gonic/gin"
)

func ProvideRouterCustomizer(handler GeneralNotificationWebSocketHandler) httpserverwrapper.RouterWithWebSocketCustomizer {
	return handler
}

func (g GeneralNotificationWebSocketHandler) Configure(b *httpserverwrapper.HTTPServerBuilder) error {
	b.WithMiddleware(gin.Recovery())
	return nil
}

func (g GeneralNotificationWebSocketHandler) RegisterRoutes(engine *gin.Engine) error {
	notificationGrouo := engine.Group("/notifications")
	{
		notificationGrouo.POST("chat", g.ForwardChatNotification)
		notificationGrouo.POST("purchase", g.ForwardPurchaseNotification)
		notificationGrouo.POST("payment-reminder", g.ForwardPaymentReminderNotification)
		notificationGrouo.POST("shipping-update", g.ForwardShippingUpdateNotification)
	}

	return nil
}

func (g GeneralNotificationWebSocketHandler) RegisterWebSocketRoutes() httpserverwrapper.WebSocketRoutes {
	return httpserverwrapper.WebSocketRoutes{
		"/notifications/subscribe": g.SubscribeNotificationWebSocketByUserID,
	}
}
