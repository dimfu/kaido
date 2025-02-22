package config

import (
	"encoding/json"
	"os"
	"path"
	"sync"

	"github.com/dimfu/kaido/models"
)

type Config struct {
	WorkspacePath     string              `json:"workspace_path"`
	KBTBaseUrl        string              `json:"kbt_base_url"`
	Leaderboards      models.Leaderboards `json:"leaderboards"`
	DiscordWebhookURL string              `json:"discord_webhook_url"`
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
	file, err := os.Create(path.Join(c.WorkspacePath, "config.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	return encoder.Encode(c)
}
