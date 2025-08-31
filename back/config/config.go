package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App    AppConfig    `yaml:"app"`
	Log    LogConfig    `yaml:"log"`
	Ollama OllamaConfig `yaml:"ollama"`
	Valkey ValkeyConfig `yaml:"valkey"`
}

type AppConfig struct {
	Env               string        `yaml:"env" validate:"required,oneof=dev staging prod production"`
	Port              int           `yaml:"port" validate:"required,min=1,max=65535"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" validate:"required"`
	ReadTimeout       time.Duration `yaml:"read_timeout" validate:"required"`
	WriteTimeout      time.Duration `yaml:"write_timeout" validate:"required"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" validate:"required"`
}

type LogConfig struct {
	Level  string `yaml:"level" validate:"required,oneof=debug info warn error"`
	Format string `yaml:"format" validate:"required,oneof=text json"`
}

type OllamaConfig struct {
	BaseURL string `yaml:"base_url" validate:"required,url"`
}

type ValkeyConfig struct {
	Address string `yaml:"address" validate:"required,hostname_port"`
}

func defaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Env:               "dev",
			Port:              8080,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		Log: LogConfig{
			Level:  "debug",
			Format: "text",
		},
		Ollama: OllamaConfig{
			BaseURL: "http://ollama:11434",
		},
		Valkey: ValkeyConfig{
			Address: "valkey:6379",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := defaultConfig()

	if fileExists(path) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	overrideFromEnv(cfg)

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func overrideFromEnv(cfg *Config) {
	overrideString("APP_ENV", &cfg.App.Env)
	overrideInt("APP_PORT", &cfg.App.Port)
	overrideDuration("APP_READ_HEADER_TIMEOUT", &cfg.App.ReadHeaderTimeout)
	overrideDuration("APP_READ_TIMEOUT", &cfg.App.ReadTimeout)
	overrideDuration("APP_WRITE_TIMEOUT", &cfg.App.WriteTimeout)
	overrideDuration("APP_IDLE_TIMEOUT", &cfg.App.IdleTimeout)

	overrideString("LOG_LEVEL", &cfg.Log.Level)
	overrideString("LOG_FORMAT", &cfg.Log.Format)

	overrideString("OLLAMA_BASE_URL", &cfg.Ollama.BaseURL)
	overrideString("VALKEY_ADDRESS", &cfg.Valkey.Address)
}

func overrideString(key string, target *string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

func overrideInt(key string, target *int) {
	if v := os.Getenv(key); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			*target = val
		}
	}
}

func overrideDuration(key string, target *time.Duration) {
	if v := os.Getenv(key); v != "" {
		if val, err := time.ParseDuration(v); err == nil {
			*target = val
		}
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
