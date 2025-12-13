package connectionconfig

import (
	"github.com/kelseyhightower/envconfig"
)

// RedisClientConfig contains Redis connection settings
type RedisClientConfig struct {
	Addr     string `envconfig:"ADDR" required:"true"`
	Password string `envconfig:"PASSWORD" default:""`
	DB       int    `envconfig:"DB" default:"0"`
	PoolSize int    `envconfig:"POOL_SIZE" default:"10"`
}

func ProvideRedisClientConfig() (conf RedisClientConfig) {
	envconfig.MustProcess("", &conf)
	return
}
