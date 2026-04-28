package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int            `yaml:"port"`
	LogLevel string         `yaml:"logLevel"`
	Calendar CalendarConfig `yaml:"calendar"`
}

type CalendarConfig struct {
	Provider string       `yaml:"provider"`
	Google   GoogleConfig `yaml:"google"`
	ICalURL  string       `yaml:"icalURL"`
}

type GoogleConfig struct {
	CredentialsFile string `yaml:"credentialsFile"`
	CalendarID      string `yaml:"calendarID"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- reading config from user-provided path is intended
	if err != nil {
		return Config{}, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.Calendar.Provider == "" {
		return fmt.Errorf("calendar.provider is required")
	}
	if c.Calendar.ICalURL == "" {
		return fmt.Errorf("calendar.icalURL is required")
	}

	switch c.Calendar.Provider {
	case "google":
		if c.Calendar.Google.CredentialsFile == "" {
			return fmt.Errorf("calendar.google.credentialsFile is required when provider is google")
		}
		if c.Calendar.Google.CalendarID == "" {
			return fmt.Errorf("calendar.google.calendarID is required when provider is google")
		}
	default:
		return fmt.Errorf("unsupported calendar provider: %s", c.Calendar.Provider)
	}

	return nil
}

func ParseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func GetConfigPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "local.yaml"
}
