package config

import (
	"os"
	"path/filepath"
	"testing"
)

func withConfigPaths(t *testing.T, searchPaths []string) {
	t.Helper()

	originalPaths := paths
	paths = searchPaths
	t.Cleanup(func() {
		paths = originalPaths
	})
}

func writeConfigFile(t *testing.T, dir, content string) {
	t.Helper()

	configPath := filepath.Join(dir, configName+"."+configFileType)
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func TestNew_LoadsConfigFromCurrentDirectory(t *testing.T) {
	withConfigPaths(t, []string{"."})

	dir := t.TempDir()
	writeConfigFile(t, dir, validSimpleDDNSYAML)

	t.Chdir(dir)

	c, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if c == nil {
		t.Fatal("New() = nil, want config")
	}

	got, err := c.GetSimpleDDNSConfig()
	if err != nil {
		t.Fatalf("GetSimpleDDNSConfig() error = %v", err)
	}
	if got.DDNS.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", got.DDNS.LogLevel, "debug")
	}
}

func TestNew_ReturnsErrorWhenConfigMissing(t *testing.T) {
	withConfigPaths(t, []string{"."})

	dir := t.TempDir()
	t.Chdir(dir)

	_, err := New()
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}
}
