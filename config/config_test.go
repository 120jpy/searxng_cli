package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigDirEnvVar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEARXNG_CLI_CONFIG_DIR", dir)
	got, err := configDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Fatalf("configDir() = %s, want %s", got, dir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEARXNG_CLI_CONFIG_DIR", dir)

	cfg := DefaultConfig()
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DefaultInstance != "local" {
		t.Fatalf("DefaultInstance = %q, want local", loaded.DefaultInstance)
	}
	if loaded.Instances["local"].URL != "http://127.0.0.1:8888" {
		t.Fatalf("URL = %q, want http://127.0.0.1:8888", loaded.Instances["local"].URL)
	}
}

func TestLoadFrom(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/custom.yaml"

	cfg := DefaultConfig()
	cfg.DefaultInstance = "custom"
	cfg.Instances["custom"] = Instance{URL: "https://example.com"}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DefaultInstance != "custom" {
		t.Fatalf("DefaultInstance = %q, want custom", loaded.DefaultInstance)
	}
}

func TestGetInstanceDefault(t *testing.T) {
	cfg := DefaultConfig()
	inst, err := cfg.GetInstance("")
	if err != nil {
		t.Fatal(err)
	}
	if inst.URL != "http://127.0.0.1:8888" {
		t.Fatalf("URL = %q, want http://127.0.0.1:8888", inst.URL)
	}
}

func TestGetInstanceByName(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Instances["other"] = Instance{URL: "https://other.com"}
	inst, err := cfg.GetInstance("other")
	if err != nil {
		t.Fatal(err)
	}
	if inst.URL != "https://other.com" {
		t.Fatalf("URL = %q, want https://other.com", inst.URL)
	}
}

func TestGetInstanceNotFound(t *testing.T) {
	cfg := DefaultConfig()
	_, err := cfg.GetInstance("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEARXNG_CLI_CONFIG_DIR", dir)
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when config missing")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DefaultInstance != "local" {
		t.Fatalf("DefaultInstance = %q", cfg.DefaultInstance)
	}
	if len(cfg.Instances) != 1 {
		t.Fatalf("got %d instances, want 1", len(cfg.Instances))
	}
}
