package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Patterns struct {
	UseDefaults      bool     `json:"use_defaults"`
	Extra            []string `json:"extra"`
	DisabledDefaults []string `json:"disabled_defaults"`
}

type Config struct {
	Sound           string   `json:"sound"`
	CustomSoundPath string   `json:"custom_sound_path"`
	Volume          float64  `json:"volume"`
	CooldownMs      int      `json:"cooldown_ms"`
	Patterns        Patterns `json:"patterns"`
}

func Default() Config {
	return Config{
		Sound:      "default",
		Volume:     0.8,
		CooldownMs: 1500,
		Patterns: Patterns{
			UseDefaults:      true,
			Extra:            []string{},
			DisabledDefaults: []string{},
		},
	}
}

func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".broask.json")
}

func Load() (Config, error) {
	cfg := Default()
	path := Path()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, Save(cfg)
	}
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func Save(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0644)
}
