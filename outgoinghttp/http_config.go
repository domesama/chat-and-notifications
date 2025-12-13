package outgoinghttp

import "time"

type OutGoingHTTPConfig struct {
	Host    string        `envconfig:"CLIENT_HOST" required:"true"`
	Timeout time.Duration `envconfig:"CLIENT_TIMEOUT" default:"2s"`
}
