package main

import (
	"path/filepath"
	"testing"
)

func TestTelegramRuntimeSelectedLabels(t *testing.T) {
	a := &app{
		accountFlag: "all",
		cfg: Config{
			Accounts: map[string]Account{
				"work": {},
			},
		},
	}

	labels, err := a.runtime().selectedLabels()
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 2 || labels[0] != defaultAccount || labels[1] != "work" {
		t.Fatalf("labels = %v, want [default work]", labels)
	}
}

func TestEnsureAccountRequiresConfiguredNamedAccount(t *testing.T) {
	a := &app{
		accountFlag: "work",
		configPath:  filepath.Join(t.TempDir(), "gotd.cli.yaml"),
		cfg:         Config{},
	}

	err := a.ensureAccount()
	if err == nil {
		t.Fatal("expected unknown named account to fail")
	}
}
