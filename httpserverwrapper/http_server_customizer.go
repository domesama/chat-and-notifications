package httpserverwrapper

import (
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/gin-gonic/gin"
)

// RouterCustomizer allows customization of HTTP server configuration and routing
type RouterCustomizer interface {
	// Configure allows modification of server builder before engine creation
	Configure(builder *HTTPServerBuilder) error

	// RegisterRoutes allows registration of routes on the Gin engine
	RegisterRoutes(engine *gin.Engine) error
}

// RouterCustomizer allows customization of HTTP server configuration and routing
type RouterWithWebSocketCustomizer interface {
	RouterCustomizer
	RegisterWebSocketRoutes() WebSocketRoutes
}

// WebSocketRoutes are maps between the route path to the route handler
type (
	WebSocketRoutes  map[string]WebsocketHandler
	WebsocketHandler func(gctx *gin.Context) (key string, metadata websocket.Metadata)
)
