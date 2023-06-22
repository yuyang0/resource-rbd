package config

import (
	"github.com/jinzhu/configor"
	coretypes "github.com/projecteru2/core/types"
)

type Config struct {
	ID        int64                `yaml:"id" required:"true"`
	Etcd      coretypes.EtcdConfig `yaml:"etcd"`
	LogLevel  string               `yaml:"log_level" required:"true" default:"INFO"`
	SentryDSN string               `yaml:"sentry_dsn"` // sentry dsn
}

// LoadConfig load config from yaml
func New(configPath string) (Config, error) {
	config := Config{}

	return config, configor.Load(&config, configPath)
}
