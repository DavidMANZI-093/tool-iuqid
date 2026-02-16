package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func FindConfigFile() string {
	locations := []string{}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		locations = append(locations, filepath.Join(xdg, "tool-iquid", "config.json"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(home, ".tool-iquid", "config.json"))
		locations = append(locations, filepath.Join(home, ".config", "tool-iquid", "config.json"))
	}

	locations = append(locations, "config.json")

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

type Config struct {
	RouterURL     string        `json:"router_url"`
	Username      string        `json:"username"`
	Password      string        `json:"password"`
	TargetSSID    string        `json:"target_ssid"`
	Cooldown      time.Duration `json:"cooldown"`
	CheckInterval time.Duration `json:"check_interval"`
	Timeout       time.Duration `json:"timeout"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.RouterURL == "" {
		config.RouterURL = "http://192.168.1.254"
	}
	if config.TargetSSID == "" {
		config.TargetSSID = "WiFi - The House"
	}
	if config.Cooldown == 0 {
		config.Cooldown = 5 * time.Minute
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &config, nil
}
