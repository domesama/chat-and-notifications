package httpserverwrapper

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/domesama/chat-and-notifications/websocket"
	gorillaws "github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
)

// HTTPWithWebSocketServer extends the HTTPServer to support WebSocket routes
type HTTPWithWebSocketServer struct {
	HTTPServer
	wsManager websocket.WebSocketManager
}

func ProvideHTTPWithWebSocketServer(
	cfg HTTPServerConfig,
	customizer RouterWithWebSocketCustomizer,
	wsManager websocket.WebSocketManager,
) (srv HTTPWithWebSocketServer, cleanUp func(), err error) {
	baseHTTPServer, cleanUp, err := newHTTPServer(cfg, customizer)
	if err != nil {
		return
	}

	webSocketServer := &HTTPWithWebSocketServer{
		HTTPServer: baseHTTPServer,
		wsManager:  wsManager,
	}

	err = webSocketServer.registerWebSocketRoutes(baseHTTPServer.engine, customizer)
	if err != nil {
		return
	}

	if err = startHTTPServerWithListener(&webSocketServer.HTTPServer); err != nil {
		return
	}

	return *webSocketServer, cleanUp, nil
}

func (s HTTPWithWebSocketServer) registerWebSocketRoutes(engine *gin.Engine,
	customizer RouterWithWebSocketCustomizer) error {

	for routePath, routeFunc := range customizer.RegisterWebSocketRoutes() {

		currentHandler := func(gctx *gin.Context) {
			key, metadata := routeFunc(gctx)
			websocketCon := s.upgradeToWebSocket(gctx)
			managedWebSocketCon := s.wsManager.RegisterConnection(key, metadata, websocketCon)
			slog.Info("registered WebSocket route", "path", routePath, "key", key, "metadata", metadata)

			// Wait for managedWebSocketCon to close
			// The manager handles heartbeat and cleanup
			<-managedWebSocketCon.CloseChan
		}

		engine.GET(routePath, func(gctx *gin.Context) { currentHandler(gctx) })
	}
	return nil
}

func (s HTTPWithWebSocketServer) upgradeToWebSocket(gctx *gin.Context) (conn *gorillaws.Conn) {
	// TODO: In the future allow implementor to add config for each socket handlers
	upgrader := gorillaws.Upgrader{}
	conn, err := upgrader.Upgrade(gctx.Writer, gctx.Request, nil)

	if err != nil {
		slog.ErrorContext(gctx.Request.Context(), "failed to upgrade WebSocket connection", "error", err)
		_ = gctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	return conn
}

func (s HTTPWithWebSocketServer) Shutdown() {
	slog.Info("shutting down HTTP server")

	// Close all WebSocket connections gracefully
	s.wsManager.CloseAll(1001, "server shutting down")

	// Give clients time to reconnect
	time.Sleep(5 * time.Second)
	s.HTTPServer.Shutdown()
}
