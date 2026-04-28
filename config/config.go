package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Instance struct {
	URL string `yaml:"url"`
}

type Config struct {
	DefaultInstance string              `yaml:"default_instance"`
	Instances       map[string]Instance `yaml:"instances"`
}

func configDir() (string, error) {
	if v := os.Getenv("SEARXNG_CLI_CONFIG_DIR"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home dir: %w", err)
	}
	return filepath.Join(home, ".searxng_cli"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	if cfg.Instances == nil {
		cfg.Instances = make(map[string]Instance)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) GetInstance(name string) (*Instance, error) {
	if name == "" {
		name = c.DefaultInstance
	}
	inst, ok := c.Instances[name]
	if !ok {
		return nil, fmt.Errorf("instance %q not found in config", name)
	}
	return &inst, nil
}

func DefaultConfig() *Config {
	return &Config{
		DefaultInstance: "local",
		Instances: map[string]Instance{
			"local": {URL: "http://127.0.0.1:8888"},
		},
	}
}
