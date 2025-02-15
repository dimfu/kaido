package config

import "sync"

type Config struct {
	WorkspacePath string
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}
