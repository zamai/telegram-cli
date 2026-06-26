package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestAccountsAddRequiresAppCredentials(t *testing.T) {
	a := &app{
		configPath: filepath.Join(t.TempDir(), "gotd.cli.yaml"),
		cfg:        Config{},
	}
	cmd := a.newAccountsAddCmd()
	cmd.SetArgs([]string{"work"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing credentials to fail")
	}
	if !strings.Contains(err.Error(), "app credentials required") {
		t.Fatalf("error = %q, want app credentials required", err)
	}
}

func TestAccountsAddWritesUserProvidedAppCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gotd.cli.yaml")
	a := &app{
		configPath: path,
		cfg:        Config{},
	}
	cmd := a.newAccountsAddCmd()
	cmd.SetArgs([]string{"work", "--app-id", "10", "--app-hash", "abcd"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	acc, err := cfg.account("work")
	if err != nil {
		t.Fatal(err)
	}
	if acc.AppID != 10 || acc.AppHash != "abcd" {
		t.Fatalf("credentials = %d/%q, want 10/abcd", acc.AppID, acc.AppHash)
	}
}
