package main

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"

	"github.com/gotd/cli/internal/output"
)

// errNotAuthorized is returned when a user-session command runs without a
// logged-in session.
var errNotAuthorized = errors.New("not authorized: run `tg login` first")

// authKind selects which session/credentials a command uses.
type authKind int

const (
	authUser authKind = iota // personal account session (default)
	authBot                  // bot token session
)

// Session kind names, used for session filenames and flags.
const (
	kindUser = "user"
	kindBot  = "bot"
)

func (k authKind) String() string {
	if k == authBot {
		return kindBot
	}
	return kindUser
}

// accountState is the resolved runtime state for the active account.
type accountState struct {
	label    string
	acc      Account
	resolver dcs.Resolver
}

// app holds shared state and the values of the global (persistent) flags.
type app struct {
	// Global flags, bound to the root command's persistent flags.
	configPath   string
	debugInvoker bool
	outputFormat string
	proxyURL     string
	accountFlag  string

	cfg     Config
	log     *zap.Logger
	waiter  *floodwait.Waiter
	printer *output.Printer

	// active is the account currently being operated on (set per run iteration).
	active *accountState

	debug bool
}

func newApp() *app {
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Level.SetLevel(zap.WarnLevel)

	defaultLog, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}

	return &app{
		waiter:  floodwait.NewWaiter(),
		log:     defaultLog,
		printer: output.New(output.Text, nil),
	}
}

// before is wired as the root command's PersistentPreRunE; it runs before every
// subcommand. It sets up the output printer and (for commands that need it)
// loads the config.
func (a *app) before(cmd *cobra.Command) error {
	format, err := output.ParseFormat(a.outputFormat)
	if err != nil {
		return err
	}
	a.printer = output.New(format, cmd.OutOrStdout())

	if skipConfig(cmd) {
		return nil
	}

	cfg, err := loadConfig(a.configPath)
	if err != nil {
		return err
	}
	a.cfg = cfg
	if a.debugInvoker {
		a.debug = true
	}
	return nil
}

// selectedLabels returns the account labels the command should run against.
func (a *app) selectedLabels() ([]string, error) {
	return a.runtime().selectedLabels()
}

// activate resolves and installs the active account state for label. When multi
// is set, the account label is included in output.
func (a *app) activate(label string, multi bool) error {
	return a.runtime().activate(label, multi)
}

// ensureActive activates a single selected account if none is active yet.
func (a *app) ensureActive() error {
	return a.runtime().ensureActive()
}

// skipConfig reports whether the command runs without a loaded config/session.
func skipConfig(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "init", "docs", "completion", "help",
			cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
			return true
		}
	}
	return false
}

// runParams configures a client session for connect/run.
type runParams struct {
	auth    authKind
	updates bool
	// authorize performs interactive authentication when the session is not yet
	// authorized. If nil, user sessions error with errNotAuthorized and bot
	// sessions authenticate with the configured token.
	authorize func(ctx context.Context, client *telegram.Client, d tg.UpdateDispatcher) error
}

// optionsFor builds telegram.Options for a specific account state.
func (a *app) optionsFor(st *accountState, rp runParams, d tg.UpdateDispatcher) telegram.Options {
	return a.runtime().optionsFor(st, rp, d)
}

// connectWith builds a client for the given account state and runs f inside the
// flood-wait + client run loop. The dispatcher is non-nil only when rp.updates
// is set.
func (a *app) connectWith(
	ctx context.Context,
	st *accountState,
	rp runParams,
	f func(ctx context.Context, client *telegram.Client, d tg.UpdateDispatcher) error,
) error {
	return a.runtime().connectWith(ctx, st, rp, f)
}

// connect builds a client for the active account and runs f. It activates a
// single selected account if none is active yet.
func (a *app) connect(
	ctx context.Context,
	rp runParams,
	f func(ctx context.Context, client *telegram.Client, d tg.UpdateDispatcher) error,
) error {
	return a.runtime().connect(ctx, rp, f)
}

// accountState resolves an account label into runtime state (without mutating
// the shared active state), for concurrent multi-account use.
func (a *app) accountState(label string) (*accountState, error) {
	return a.runtime().accountState(label)
}

// run connects, ensures the session is authorized, and calls f with the API
// client, once per selected account. With --account all it fans out across all
// configured accounts (sequentially), labeling each result. User sessions must
// already be logged in (unless rp.authorize is set); bot sessions authenticate
// with the configured token on demand.
func (a *app) run(
	ctx context.Context,
	rp runParams,
	f func(ctx context.Context, api *tg.Client) error,
) error {
	return a.runtime().run(ctx, rp, f)
}

// runOne authorizes and runs f against the active account.
func (a *app) runOne(
	ctx context.Context,
	rp runParams,
	f func(ctx context.Context, api *tg.Client) error,
) error {
	return a.runtime().runOne(ctx, rp, f)
}
