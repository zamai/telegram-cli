package main

import (
	"context"
	"path/filepath"

	"github.com/go-faster/errors"

	"github.com/gotd/log/logzap"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"

	"github.com/gotd/cli/internal/pretty"
	"github.com/gotd/cli/internal/proxy"
)

// telegramRuntime owns account/session/client execution decisions for app.
type telegramRuntime struct {
	app *app
}

func (a *app) runtime() telegramRuntime {
	return telegramRuntime{app: a}
}

// selectedLabels returns the account labels the command should run against.
func (r telegramRuntime) selectedLabels() ([]string, error) {
	a := r.app
	if a.accountFlag == "all" {
		labels := a.cfg.labels()
		if len(labels) == 0 {
			return nil, errors.New("no configured accounts")
		}
		return labels, nil
	}
	label := a.accountFlag
	if label == "" {
		// No --account / TG_ACCOUNT: use the configured default account.
		label = a.cfg.resolvedDefault()
	}
	return []string{label}, nil
}

func (r telegramRuntime) activate(label string, multi bool) error {
	a := r.app
	st, err := r.accountState(label)
	if err != nil {
		return err
	}
	a.active = st
	if multi {
		a.printer.SetAccount(label)
	}
	return nil
}

func (r telegramRuntime) ensureActive() error {
	a := r.app
	if a.active != nil {
		return nil
	}
	labels, err := r.selectedLabels()
	if err != nil {
		return err
	}
	if len(labels) != 1 {
		return errors.New("this command needs a single --account (not 'all')")
	}
	return r.activate(labels[0], false)
}

func (r telegramRuntime) accountState(label string) (*accountState, error) {
	a := r.app
	acc, err := a.cfg.account(label)
	if err != nil {
		return nil, err
	}
	proxyURL := a.proxyURL
	if proxyURL == "" {
		proxyURL = acc.Proxy
	}
	resolver, err := proxy.Resolver(proxyURL)
	if err != nil {
		return nil, err
	}
	if resolver == nil {
		// No proxy: connect like Telegram Desktop (Obfuscated2 + abridged
		// transport) instead of gotd's plain default.
		resolver = telegram.TDesktopResolver()
	}
	return &accountState{label: label, acc: acc, resolver: resolver}, nil
}

func (r telegramRuntime) optionsFor(st *accountState, rp runParams, d tg.UpdateDispatcher) telegram.Options {
	a := r.app
	mw := []telegram.Middleware{a.waiter}
	if a.debug {
		mw = append(mw, pretty.Middleware())
	}

	opts := telegram.Options{
		Logger:      logzap.New(a.log.Named("tg")),
		Device:      deviceConfig(),
		Middlewares: mw,
		SessionStorage: &session.FileStorage{
			Path: st.acc.sessionPath(filepath.Dir(a.configPath), st.label, rp.auth.String()),
		},
	}
	if rp.updates {
		opts.UpdateHandler = d
	} else {
		opts.NoUpdates = true
	}
	if st.acc.Test {
		opts.DCList = dcs.Test()
	}
	opts.Resolver = st.resolver
	return opts
}

func (r telegramRuntime) connectWith(
	ctx context.Context,
	st *accountState,
	rp runParams,
	f func(ctx context.Context, client *telegram.Client, d tg.UpdateDispatcher) error,
) error {
	a := r.app
	var d tg.UpdateDispatcher
	if rp.updates {
		d = tg.NewUpdateDispatcher()
	}

	appID, appHash, err := effectiveCreds(st.acc)
	if err != nil {
		return err
	}
	client := telegram.NewClient(appID, appHash, r.optionsFor(st, rp, d))

	if err := a.waiter.Run(ctx, func(ctx context.Context) error {
		return client.Run(ctx, func(ctx context.Context) error {
			return f(ctx, client, d)
		})
	}); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (r telegramRuntime) connect(
	ctx context.Context,
	rp runParams,
	f func(ctx context.Context, client *telegram.Client, d tg.UpdateDispatcher) error,
) error {
	if err := r.ensureActive(); err != nil {
		return err
	}
	return r.connectWith(ctx, r.app.active, rp, f)
}

func (r telegramRuntime) run(
	ctx context.Context,
	rp runParams,
	f func(ctx context.Context, api *tg.Client) error,
) error {
	a := r.app
	labels, err := r.selectedLabels()
	if err != nil {
		return err
	}
	multi := len(labels) > 1

	for _, label := range labels {
		a.active = nil
		if err := r.activate(label, multi); err != nil {
			return err
		}
		if err := r.runOne(ctx, rp, f); err != nil {
			if multi {
				return errors.Wrapf(err, "account %q", label)
			}
			return err
		}
	}
	return nil
}

func (r telegramRuntime) runOne(
	ctx context.Context,
	rp runParams,
	f func(ctx context.Context, api *tg.Client) error,
) error {
	a := r.app
	return r.connect(ctx, rp, func(ctx context.Context, client *telegram.Client, d tg.UpdateDispatcher) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return errors.Wrap(err, "auth status")
		}
		if !status.Authorized {
			switch {
			case rp.authorize != nil:
				if err := rp.authorize(ctx, client, d); err != nil {
					return err
				}
			case rp.auth == authBot:
				if a.active.acc.BotToken == "" {
					return errors.New("no bot_token in config")
				}
				if _, err := client.Auth().Bot(ctx, a.active.acc.BotToken); err != nil {
					return errors.Wrap(err, "bot auth")
				}
			default:
				return errNotAuthorized
			}
		}
		return f(ctx, client.API())
	})
}
