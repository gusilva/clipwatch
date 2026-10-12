package main

import (
	"os"
	"os/exec"
)

func openPopup() {
	herdr := os.Getenv("HERDR_BIN_PATH")
	if herdr == "" {
		fatal("open must be run by herdr (HERDR_BIN_PATH is not set)")
	}

	cmd := exec.Command(
		herdr,
		"plugin",
		"pane",
		"open",
		"--plugin",
		os.Getenv("HERDR_PLUGIN_ID"),
		"--entrypoint",
		"picker",
	)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fatal("error", "err", err)
	}
}
