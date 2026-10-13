package main

import (
	"os"
	"path/filepath"
	"time"
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

// heartbeatEvery is how often the watcher touches its heartbeat file.
const heartbeatEvery = time.Second

// heartbeatMaxAge is how old the heartbeat file may be before the watcher
// counts as stopped. It allows for a few missed beats.
const heartbeatMaxAge = 3 * heartbeatEvery

// heartbeatPath returns the path of the heartbeat file inside dataDir. The
// watcher touches it while it runs, and the picker reads its modification time.
func heartbeatPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "heartbeat"), nil
}

// touchHeartbeat creates the heartbeat file if needed and sets its
// modification time to now.
func touchHeartbeat(path string, now time.Time) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return os.Chtimes(path, now, now)
}

// watcherAlive reports whether the heartbeat file at path was touched within
// heartbeatMaxAge before now. It does not depend on launchd or its output
// format, and it is also false for a watcher that hangs. An empty path or a
// missing file means the watcher is not running.
func watcherAlive(path string, now time.Time) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return now.Sub(info.ModTime()) <= heartbeatMaxAge
}
