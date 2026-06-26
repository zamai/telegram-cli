package main

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram"
)

// effectiveCreds resolves the app id/hash for an account. The test server uses
// these same user-provided credentials too; only the DC list differs (see
// optionsFor), so the gotd test credentials are not used.
func effectiveCreds(acc Account) (appID int, appHash string, err error) {
	if err := requireAppCredentials(acc.AppID, acc.AppHash); err != nil {
		return 0, "", err
	}
	return acc.AppID, acc.AppHash, nil
}

func requireAppCredentials(appID int, appHash string) error {
	if appID != 0 && appHash != "" {
		return nil
	}
	return errors.New("app credentials required: pass --app-id/--app-hash to `tg init` " +
		"or `tg accounts add` (get them from https://my.telegram.org)")
}

// deviceConfig mimics Telegram Desktop (Windows) so the session shows up as a
// desktop client in Settings → Devices. It delegates to gotd's built-in preset,
// which keeps the app version in sync with the bundled tdesktop reference and
// sends the same initConnection params (including the tz_offset timezone field)
// as the real client, making the connection indistinguishable from Telegram
// Desktop's. Pair it with telegram.TDesktopResolver on the transport layer.
func deviceConfig() telegram.DeviceConfig {
	return telegram.DeviceTDesktopWindows()
}
