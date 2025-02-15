package config

import (
	"encoding/json"
	"os"
	"path"
	"sync"

	"github.com/dimfu/kaido/models"
)

type Config struct {
	WorkspacePath string              `json:"workspace_path"`
	KBTBaseUrl    string              `json:"kbt_base_url"`
	Leaderboards  models.Leaderboards `json:"leaderboards"`
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

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(c.WorkspacePath, "config.json"), data, 0644)
}
