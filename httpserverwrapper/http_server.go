package httpserverwrapper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/gin-gonic/gin"
)

// HTTPServer wraps the Gin engine and HTTP server
type HTTPServer struct {
	engine    *gin.Engine
	server    *http.Server
	listener  net.Listener
	wsManager websocket.WebSocketManager
	cfg       HTTPServerConfig
}

func ProvideHTTPServer(
	cfg HTTPServerConfig,
	customizer RouterCustomizer,
) (srv HTTPServer, cleanUp func(), err error) {
	srv, cleanUp, err = newHTTPServer(cfg, customizer)
	if err != nil {
		return
	}

	if err = startHTTPServerWithListener(&srv); err != nil {
		return HTTPServer{}, func() {}, err
	}

	return
}

// newHTTPServer creates a new HTTP server with router customizer pattern
func newHTTPServer(
	cfg HTTPServerConfig,
	customizer RouterCustomizer,
) (srv HTTPServer, cleanUp func(), err error) {
	builder := NewHTTPServerBuilder(cfg)

	if err = customizer.Configure(builder); err != nil {
		return
	}

	engine := builder.Build()

	if err := customizer.RegisterRoutes(engine); err != nil {
		return HTTPServer{}, func() {}, fmt.Errorf("failed to register routes: %w", err)
	}

	server := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	httpServer := HTTPServer{
		engine: engine,
		server: server,
		cfg:    cfg,
	}

	cleanup := func() {
		httpServer.Shutdown()
	}

	return httpServer, cleanup, nil
}

// Shutdown gracefully shuts down the HTTP server
func (s HTTPServer) Shutdown() {
	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}
}

func (s HTTPServer) GetRunningPort() string {
	if s.listener != nil {
		_, port, err := net.SplitHostPort(s.listener.Addr().String())
		if err == nil {
			return ":" + port
		}
	}
	return s.server.Addr
}

func startHTTPServerWithListener(srv *HTTPServer) error {
	listener, err := net.Listen("tcp", srv.server.Addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	srv.listener = listener

	go func() {
		actualAddr := listener.Addr().String()
		slog.Info("Starting HTTP server", "addr", actualAddr)
		if err := srv.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	return nil
}
