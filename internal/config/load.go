package config

import (
	"encoding/json/v2"
	"fmt"
	"os"
)

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ReleaseDir:   "releases",
		UploadDir:    "uploads",
		UnpackScript: "scripts/unpack.sh",
		ReloadScript: "scripts/reload.sh",
	}

	data, err := os.ReadFile("config.json")

	if err != nil {
		return nil, fmt.Errorf("Failed to read config file: %w", err)
	}

	err = json.Unmarshal(data, cfg)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
