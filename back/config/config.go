package config

import (
	"fmt"
	"os"
	"strconv"

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
	Env  string `yaml:"env" validate:"required,oneof=dev staging prod production"`
	Port int    `yaml:"port" validate:"required,min=1,max=65535"`
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
			Env:  "dev",
			Port: 8080,
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
