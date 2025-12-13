package ittesthelper

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type BeforeServerStartFn func(*gin.Engine) error

func StartHTTPServer(t *testing.T, beforeServerStartFn BeforeServerStartFn) int {

	addr := "localhost:0"

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	router := gin.New()

	router.GET(
		"/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{})
		},
	)

	if err := beforeServerStartFn(router); err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           router.Handler(),
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
	}
	ctx := context.Background()
	t.Cleanup(
		func() {
			if err := server.Shutdown(ctx); err != nil {
				slog.Error("cannot shutdown http server", err)
			}
		},
	)

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("caught error during serving http")
		}
	}()

	listenerPort := listener.Addr().(*net.TCPAddr).Port

	client := NewTestHTTPClient(t, BaseLocalURL(listenerPort)+"/")
	client.HealthCheck(client.CreateJSONRequest(context.Background(), "GET", "health", "{}"))

	return listenerPort
}

func BaseLocalURL(port int) string {
	return "http://localhost:" + strconv.Itoa(port)
}
