package main

import (
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals // Release builds stamp these with -ldflags -X.
var (
	buildVersion = "dev"
	buildCommit  = ""
	buildDate    = ""
)

type versionResult struct {
	Version       string `json:"version"`
	Commit        string `json:"commit,omitempty"`
	BuildDate     string `json:"build_date,omitempty"`
	VCSDate       string `json:"vcs_date,omitempty"`
	Dirty         bool   `json:"dirty,omitempty"`
	GoVersion     string `json:"go_version"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
	Compiler      string `json:"compiler"`
	ModulePath    string `json:"module_path,omitempty"`
	GotdTDVersion string `json:"gotd_td_version,omitempty"`
}

func newVersionResult() versionResult {
	info, ok := debug.ReadBuildInfo()
	r := versionResult{
		Version:   buildVersion,
		Commit:    buildCommit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		Compiler:  runtime.Compiler,
	}
	if !ok {
		return r
	}

	r.ModulePath = info.Main.Path
	r.GotdTDVersion = depVersion(info.Deps, "github.com/gotd/td")
	settings := buildSettings(info.Settings)
	if r.Commit == "" {
		r.Commit = settings["vcs.revision"]
	}
	r.VCSDate = settings["vcs.time"]
	r.Dirty = settings["vcs.modified"] == "true"
	if r.Version == "" || r.Version == "dev" {
		switch {
		case info.Main.Version != "" && info.Main.Version != "(devel)":
			r.Version = info.Main.Version
		case r.Commit != "":
			r.Version = r.Commit
		default:
			r.Version = "dev"
		}
	}
	return r
}

func depVersion(deps []*debug.Module, path string) string {
	for _, dep := range deps {
		if dep.Path == path {
			if dep.Replace != nil {
				return dep.Replace.Version
			}
			return dep.Version
		}
	}
	return ""
}

func buildSettings(settings []debug.BuildSetting) map[string]string {
	out := make(map[string]string, len(settings))
	for _, setting := range settings {
		out[setting.Key] = setting.Value
	}
	return out
}

func (r versionResult) MarshalText(w io.Writer) error {
	lines := []string{
		"version: " + r.Version,
		"go: " + r.GoVersion,
		"platform: " + r.GOOS + "/" + r.GOARCH,
		"compiler: " + r.Compiler,
	}
	if r.Commit != "" {
		lines = append(lines, "commit: "+r.Commit)
	}
	if r.BuildDate != "" {
		lines = append(lines, "build_date: "+r.BuildDate)
	}
	if r.VCSDate != "" {
		lines = append(lines, "vcs_date: "+r.VCSDate)
	}
	if r.Dirty {
		lines = append(lines, "dirty: true")
	}
	if r.ModulePath != "" {
		lines = append(lines, "module: "+r.ModulePath)
	}
	if r.GotdTDVersion != "" {
		lines = append(lines, "gotd_td: "+r.GotdTDVersion)
	}
	_, err := fmt.Fprintln(w, strings.Join(lines, "\n"))
	return err
}

func (a *app) newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   cmdVersion,
		Short: "Show build and platform information",
		Long: `Print information about the running tg binary: release version or commit,
build date when stamped by the release build, Go version, platform, and VCS metadata.`,
		Example: `  tg version
  tg version --output json`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return a.printer.Emit(newVersionResult())
		},
	}
}
