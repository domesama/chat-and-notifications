package websocket

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// WebSocketConfig contains WebSocket-specific settings
type WebSocketConfig struct {
	PingInterval time.Duration `envconfig:"PING_INTERVAL" default:"30s"`
	PongWait     time.Duration `envconfig:"PONG_WAIT" default:"40s"`
	WriteWait    time.Duration `envconfig:"WRITE_WAIT" default:"10s"`
}

func ProvideWebSocketConfig() (conf WebSocketConfig) {
	envconfig.MustProcess("", &conf)
	return
}
