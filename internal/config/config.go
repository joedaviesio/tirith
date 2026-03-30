package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy     ProxyConfig     `yaml:"proxy"`
	Providers ProvidersConfig `yaml:"providers"`
	Storage   StorageConfig   `yaml:"storage"`
	Dashboard DashboardConfig `yaml:"dashboard"`
	Cloud     CloudConfig     `yaml:"cloud"`
}

type ProxyConfig struct {
	Port      int  `yaml:"port"`
	LogBodies bool `yaml:"log_bodies"`
}

type ProvidersConfig struct {
	Anthropic ProviderEntry `yaml:"anthropic"`
	OpenAI    ProviderEntry `yaml:"openai"`
	Google    ProviderEntry `yaml:"google"`
}

type ProviderEntry struct {
	Upstream string `yaml:"upstream"`
}

type StorageConfig struct {
	Driver string `yaml:"driver"`
	Path   string `yaml:"path"`
}

type DashboardConfig struct {
	Port int `yaml:"port"`
}

type CloudConfig struct {
	Enabled  bool   `yaml:"enabled"`
	APIKey   string `yaml:"api_key"`
	Endpoint string `yaml:"endpoint"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Proxy: ProxyConfig{
			Port:      5555,
			LogBodies: false,
		},
		Providers: ProvidersConfig{
			Anthropic: ProviderEntry{Upstream: "https://api.anthropic.com"},
			OpenAI:    ProviderEntry{Upstream: "https://api.openai.com"},
			Google:    ProviderEntry{Upstream: "https://generativelanguage.googleapis.com"},
		},
		Storage: StorageConfig{
			Driver: "sqlite",
			Path:   filepath.Join(home, ".costwatch", "data.db"),
		},
		Dashboard: DashboardConfig{
			Port: 5556,
		},
		Cloud: CloudConfig{
			Enabled:  false,
			Endpoint: "https://api.costwatch.dev",
		},
	}
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".costwatch")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func EnsureDir() error {
	return os.MkdirAll(ConfigDir(), 0o755)
}

func Save(cfg *Config) error {
	if err := EnsureDir(); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(ConfigPath(), data, 0o644)
}
