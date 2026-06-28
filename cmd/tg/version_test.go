package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVersionResultMarshalText(t *testing.T) {
	res := versionResult{
		Version:       "v1.2.3",
		Commit:        "abc123",
		BuildDate:     "2026-06-28T12:00:00Z",
		VCSDate:       "2026-06-28T11:00:00Z",
		Dirty:         true,
		GoVersion:     "go1.25.0",
		GOOS:          "darwin",
		GOARCH:        "arm64",
		Compiler:      "gc",
		ModulePath:    "github.com/gotd/cli",
		GotdTDVersion: "v0.159.0",
	}

	var buf bytes.Buffer
	if err := res.MarshalText(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"version: v1.2.3",
		"commit: abc123",
		"build_date: 2026-06-28T12:00:00Z",
		"vcs_date: 2026-06-28T11:00:00Z",
		"dirty: true",
		"go: go1.25.0",
		"platform: darwin/arm64",
		"compiler: gc",
		"module: github.com/gotd/cli",
		"gotd_td: v0.159.0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestVersionCommandSkipsConfigAndEmitsJSON(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--config", filepath.Join(t.TempDir(), "missing.yaml"),
		"--output", "json",
		cmdVersion,
	})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var env struct {
		Schema int           `json:"schema"`
		Data   versionResult `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal %q: %v", out.String(), err)
	}
	if env.Schema != 1 {
		t.Errorf("schema = %d, want 1", env.Schema)
	}
	if env.Data.Version == "" {
		t.Error("version is empty")
	}
	if env.Data.GoVersion != runtime.Version() {
		t.Errorf("go_version = %q, want %q", env.Data.GoVersion, runtime.Version())
	}
	if env.Data.GOOS != runtime.GOOS || env.Data.GOARCH != runtime.GOARCH {
		t.Errorf("platform = %s/%s, want %s/%s", env.Data.GOOS, env.Data.GOARCH, runtime.GOOS, runtime.GOARCH)
	}
}
