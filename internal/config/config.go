package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const fileName = "config.json"
const defaultTabWidth = 4

type Config struct {
	NotesPath string `json:"notes_path"`
	TabWidth  int    `json:"tab_width"`
}

func LoadOrCreate() (Config, string, error) {
	configPath, err := path()
	if err != nil {
		return Config{}, "", err
	}

	if _, err := os.Stat(configPath); err == nil {
		cfg, err := load(configPath)
		return cfg, configPath, err
	} else if !os.IsNotExist(err) {
		return Config{}, "", fmt.Errorf("stat config: %w", err)
	}

	cfg, err := defaultConfig()
	if err != nil {
		return Config{}, "", err
	}

	if err := save(configPath, cfg); err != nil {
		return Config{}, "", err
	}

	return cfg, configPath, nil
}

func path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate user home dir: %w", err)
	}

	return filepath.Join(home, ".config", "nocte", fileName), nil
}

func load(configPath string) (Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if strings.TrimSpace(cfg.NotesPath) == "" {
		defaultCfg, err := defaultConfig()
		if err != nil {
			return Config{}, err
		}
		cfg.NotesPath = defaultCfg.NotesPath

		if err := save(configPath, cfg); err != nil {
			return Config{}, err
		}
	}

	if cfg.TabWidth <= 0 {
		cfg.TabWidth = defaultTabWidth

		if err := save(configPath, cfg); err != nil {
			return Config{}, err
		}
	}

	cfg.NotesPath, err = expandHome(cfg.NotesPath)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func save(configPath string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func defaultConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("locate user home dir: %w", err)
	}

	return Config{
		NotesPath: filepath.Join(home, "nocte"),
		TabWidth:  defaultTabWidth,
	}, nil
}

func expandHome(value string) (string, error) {
	if value == "~" || strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate user home dir: %w", err)
		}

		if value == "~" {
			return home, nil
		}

		return filepath.Join(home, strings.TrimPrefix(value, "~/")), nil
	}

	return value, nil
}
