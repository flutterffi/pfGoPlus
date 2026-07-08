package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithOptionsMergesProfileConfig(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, filepath.Join(dir, "config.yaml"), `
app:
  name: pfGoPlus
  env: local
database:
  dsn: ./tmp/base.db
todo_backend:
  mode: local
observability:
  exporter: stdout
`)
	writeTestFile(t, filepath.Join(dir, "config.docker.yaml"), `
app:
  env: docker
database:
  dsn: /app/tmp/docker.db
todo_backend:
  mode: grpc
observability:
  exporter: otlp
`)

	cfg, err := loadWithOptions([]string{dir}, "config", "", "docker")
	if err != nil {
		t.Fatalf("load config with profile: %v", err)
	}
	if cfg.App.Env != "docker" {
		t.Fatalf("expected docker env, got %s", cfg.App.Env)
	}
	if cfg.Database.DSN != "/app/tmp/docker.db" {
		t.Fatalf("expected docker dsn, got %s", cfg.Database.DSN)
	}
	if cfg.TodoBackend.Mode != "grpc" {
		t.Fatalf("expected grpc todo backend, got %s", cfg.TodoBackend.Mode)
	}
	if cfg.Observability.Exporter != "otlp" {
		t.Fatalf("expected otlp exporter, got %s", cfg.Observability.Exporter)
	}
}

func TestLoadWithOptionsUsesDefaultsWithoutFiles(t *testing.T) {
	cfg, err := loadWithOptions([]string{t.TempDir()}, "config", "", "")
	if err != nil {
		t.Fatalf("load defaults: %v", err)
	}
	if cfg.App.Name != "pfGoPlus" {
		t.Fatalf("expected default app name, got %s", cfg.App.Name)
	}
	if cfg.HTTP.Port != 8080 {
		t.Fatalf("expected default http port 8080, got %d", cfg.HTTP.Port)
	}
	if cfg.Observability.OTLPEndpoint != "127.0.0.1:4317" {
		t.Fatalf("expected default otlp endpoint, got %s", cfg.Observability.OTLPEndpoint)
	}
}

func TestLoadWithOptionsRespectsExplicitConfigFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "custom.yaml")

	writeTestFile(t, configFile, `
app:
  name: pfGoPlus-custom
http:
  port: 18080
`)

	cfg, err := loadWithOptions([]string{dir}, "config", configFile, "")
	if err != nil {
		t.Fatalf("load explicit config file: %v", err)
	}
	if cfg.App.Name != "pfGoPlus-custom" {
		t.Fatalf("expected custom app name, got %s", cfg.App.Name)
	}
	if cfg.HTTP.Port != 18080 {
		t.Fatalf("expected custom port 18080, got %d", cfg.HTTP.Port)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
