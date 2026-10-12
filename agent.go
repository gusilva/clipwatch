package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// watcherLabel is the launchd label of the watcher. It must match the Label
// key in app/gusilva.clipwatch.watcher.plist.
const watcherLabel = "gusilva.clipwatch.watcher"

// installAgent makes sure macOS runs the watcher, and then returns, because
// herdr hooks are one-shot. It restarts a registered watcher so that a rebuilt
// binary takes over, registers an unregistered one, and refuses to continue
// when the user has switched the watcher off in System Settings, which it
// opens. It runs as herdr's [[startup]] hook or by hand.
func installAgent() {
	switch st := watcherStatus(); st {
	case enabled:
		target := fmt.Sprintf("gui/%d/%s", os.Getuid(), watcherLabel)
		if out, err := exec.Command("launchctl", "kickstart", "-k", target).CombinedOutput(); err != nil {
			fatal("kickstart failed", "err", err, "output", strings.TrimSpace(string(out)))
		}
		slog.Info("watcher restarted", "status", st)
	case requiresApproval:
		openLoginItems()
		fatal("watcher is switched off: enable Clipwatch in System Settings → Login Items & Extensions")
	default:
		if err := registerWatcher(); err != nil {
			fatal("register failed (is this binary inside Clipwatch.app?)", "err", err)
		}
		slog.Info("watcher registered", "status", watcherStatus())
	}
}

// removeAgent unregisters the watcher, so macOS stops it and removes it from
// System Settings. It is the clean-uninstall step.
func removeAgent() {
	if err := unregisterWatcher(); err != nil {
		fatal("unregister failed", "err", err)
	}
	slog.Info("watcher removed", "status", watcherStatus())
}
