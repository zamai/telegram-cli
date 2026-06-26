package main

import "testing"

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
