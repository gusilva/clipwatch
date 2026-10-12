package main

import (
	"os"
	"path/filepath"
)

// dataDir returns the folder where clipwatch keeps its files, which is
// clipwatch inside os.UserConfigDir (~/Library/Application Support on macOS).
// The watcher is started by macOS and never sees HERDR_PLUGIN_STATE_DIR, so
// every subcommand uses this one folder.
func dataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "clipwatch"), nil
}

// historyPath returns the path of history.json inside dataDir.
func historyPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "history.json"), nil
}

// openLog creates dataDir if needed and opens watch.log inside it for
// appending. The file is created with mode 0600, so only the user can read it.
func openLog() (*os.File, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(dir, "watch.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
}
