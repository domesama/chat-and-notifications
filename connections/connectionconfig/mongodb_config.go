package connectionconfig

import "github.com/kelseyhightower/envconfig"

// MongoDBConfig contains MongoDB connection settings
type MongoDBConfig struct {
	URI      string `envconfig:"MONGO_URI" required:"true"`
	Database string `envconfig:"MONGO_DATABASE" required:"true"`
}

func ProvideMongoDBConfig() (conf MongoDBConfig) {
	envconfig.MustProcess("", &conf)
	return
}
