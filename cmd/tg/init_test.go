package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitRequiresAppCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gotd.cli.yaml")
	a := &app{configPath: path}
	cmd := newInitCmd(a)
	cmd.SetArgs(nil)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing credentials to fail")
	}
	if !strings.Contains(err.Error(), "app credentials required") {
		t.Fatalf("error = %q, want app credentials required", err)
	}
}

func TestInitWritesUserProvidedAppCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gotd.cli.yaml")
	a := &app{configPath: path}
	cmd := newInitCmd(a)
	cmd.SetArgs([]string{"--app-id", "10", "--app-hash", "abcd"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppID != 10 || cfg.AppHash != "abcd" {
		t.Fatalf("credentials = %d/%q, want 10/abcd", cfg.AppID, cfg.AppHash)
	}
}
