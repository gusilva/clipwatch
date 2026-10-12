package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Entry struct {
	Text     string    `json:"text"`
	CopiedAt time.Time `json:"copied_at"`
}

type History struct {
	Entries []Entry `json:"entries"`
}

var MaxEntries = 100

// MaxEntryBytes is the largest clip we record (64 KB). Bigger clips are skipped.
const MaxEntryBytes = 64 * 1024

func (h *History) Add(text string, now time.Time) bool {
	if len(text) > MaxEntryBytes || strings.TrimSpace(text) == "" {
		return false
	}

	for i, e := range h.Entries {
		if e.Text == text {
			h.Entries = append(h.Entries[:i], h.Entries[i+1:]...) // remoce old position
			break
		}
	}

	h.Entries = append([]Entry{{Text: text, CopiedAt: now}}, h.Entries...)
	if len(h.Entries) > MaxEntries {
		h.Entries = h.Entries[:MaxEntries]
	}

	return true
}

func (h *History) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".history-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(h); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), path)
}

func Load(path string) (*History, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &History{}, nil
	}

	if err != nil {
		return nil, err
	}

	var h History
	if err = json.Unmarshal(b, &h); err != nil {
		return nil, err
	}

	return &h, nil
}
