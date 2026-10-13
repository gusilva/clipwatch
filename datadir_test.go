package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherAlive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "heartbeat")
	now := time.Now()

	if watcherAlive(path, now) {
		t.Fatal("missing file: want not alive")
	}

	if watcherAlive("", now) {
		t.Fatal("empty path: want not alive")
	}

	if err := touchHeartbeat(path, now); err != nil {
		t.Fatal(err)
	}

	if !watcherAlive(path, now.Add(heartbeatMaxAge)) {
		t.Fatal("fresh beat: want alive")
	}

	if watcherAlive(path, now.Add(heartbeatMaxAge+time.Second)) {
		t.Fatal("stale beat: want not alive")
	}
}
