package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func texts(h *History) []string {
	out := make([]string, len(h.Entries))
	for i, e := range h.Entries {
		out[i] = e.Text
	}

	return out
}

func TestAddNewestFirst(t *testing.T) {
	var h History
	h.Add("hello", t0)
	h.Add("world", t0.Add(time.Second))

	if got := fmt.Sprint(texts(&h)); got != "[world hello]" {
		t.Fatalf("got %s, want [world hello]", got)
	}
}

func TestAddDuplicateMovesToTop(t *testing.T) {
	var h History
	h.Add("hello", t0)
	h.Add("world", t0.Add(time.Second))
	h.Add("hello", t0.Add(2*time.Second))

	if got := fmt.Sprint(texts(&h)); got != "[hello world]" {
		t.Fatalf("got %s, want [hello world]", got)
	}

	if !h.Entries[0].CopiedAt.Equal(t0.Add(2 * time.Second)) {
		t.Fatalf("moved entry should get the new timestamp, got %v", h.Entries[0].CopiedAt)
	}

}

func TestAddIgnoresBlank(t *testing.T) {
	var h History
	for _, s := range []string{"", "   ", "\n\t"} {
		if h.Add(s, t0) {
			t.Errorf("Add(%q) returned true; blank text should be ignored", s)
		}
	}
	if len(h.Entries) != 0 {
		t.Fatalf("want empty history, got %d entries", len(h.Entries))
	}
}

func TestAddCapsAtMax(t *testing.T) {
	var h History
	for i := 1; i <= MaxEntries+5; i++ {
		h.Add(fmt.Sprint("clip ", i), t0.Add(time.Duration(i)*time.Second))
	}

	if len(h.Entries) != MaxEntries {
		t.Fatalf("want %d entries, got %d", MaxEntries, len(h.Entries))
	}

	if h.Entries[0].Text != "clip 105" {
		t.Errorf("newest should be clip 105, got %q", h.Entries[0].Text)
	}

	if last := h.Entries[MaxEntries-1].Text; last != "clip 6" {
		t.Errorf("oldest kept should be clip 6 (1-5 dropped), got %q", last)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	var (
		h    History
		path = filepath.Join(t.TempDir(), "sub", "history.json")
	)

	h.Add("hello", t0)
	h.Add("line one\nline two", t0.Add(time.Second))
	if err := h.Save(path); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(texts(got)) != fmt.Sprint(texts(&h)) {
		t.Fatalf("round trip mismatch: %q vs %q", texts(got), texts(&h))
	}
}

func TestLoadMissingFileIsEmpty(t *testing.T) {
	h, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("missing file should not be an error, got %v", err)
	}

	if len(h.Entries) != 0 {
		t.Fatalf("want empty history, got %d entries", len(h.Entries))
	}
}

func TestSaveIsPrivate(t *testing.T) {
	var (
		h    History
		path = filepath.Join(t.TempDir(), "history.json")
	)

	h.Add("secret-ish", t0)
	if err := h.Save(path); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("history file should be 0600 (owner only), got %o", perm)
	}
}

func TestSaveLeavesNoTempFiles(t *testing.T) {
	var (
		h   History
		dir = t.TempDir()
	)

	h.Add("x", t0)
	if err := h.Save(filepath.Join(dir, "history.json")); err != nil {
		t.Fatal(err)
	}

	files, _ := os.ReadDir(dir)
	if len(files) != 1 {
		t.Fatalf("want only history.json in dir, found %d files", len(files))
	}
}

func TestAddSkipsHugeText(t *testing.T) {
	var h History
	h.Add("small", t0)

	huge := strings.Repeat("x", MaxEntryBytes+1)
	if h.Add(huge, t0.Add(time.Second)) {
		t.Fatal("Add returned true for text over MaxEntryBytes")
	}
	if got := fmt.Sprint(texts(&h)); got != "[small]" {
		t.Fatalf("history should be unchanged, got %d entries", len(h.Entries))
	}

	exact := strings.Repeat("y", MaxEntryBytes)
	if !h.Add(exact, t0.Add(2*time.Second)) {
		t.Fatal("text of exactly MaxEntryBytes should be accepted")
	}
}
