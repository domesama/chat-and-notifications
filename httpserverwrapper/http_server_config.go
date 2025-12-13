package httpserverwrapper

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// HTTPServerConfig contains HTTP server settings
type HTTPServerConfig struct {
	ListenAddr      string        `envconfig:"LISTEN_ADDR" default:":8080"`
	ReadTimeout     time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
	WriteTimeout    time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}

func ProvideHTTPConfig() (conf HTTPServerConfig) {
	envconfig.MustProcess("", &conf)
	return
}
