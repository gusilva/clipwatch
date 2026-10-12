package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func pickerWith(texts ...string) model {
	m := newModel("")
	for _, s := range texts {
		m.entries = append(m.entries, Entry{Text: s})
	}
	return m
}

// typeText sends each character as a key press, like a person typing.
func typeText(m model, s string) model {
	for _, r := range s {
		m, _ = send(m, key(r))
	}
	return m
}

func visibleTexts(m model) []string {
	var out []string
	for _, e := range m.visible() {
		out = append(out, e.Text)
	}
	return out
}

func TestSlashEntersFilterMode(t *testing.T) {
	m, _ := send(pickerWith("alpha", "beta"), key('/'))
	if !m.filtering {
		t.Fatal("/ should switch to filtering mode")
	}
}

func TestJTypesWhileFiltering(t *testing.T) {
	m, _ := send(pickerWith("alpha", "jq '.x'", "jest run"), key('/'))
	m = typeText(m, "j")
	if got := m.query.Value(); got != "j" {
		t.Fatalf("while filtering, j must be typed: want query \"j\", got %q", got)
	}
	if m.cursor != 0 {
		t.Fatalf("while filtering, j must NOT move the cursor: got %d", m.cursor)
	}
}

func TestJMovesWhileBrowsing(t *testing.T) {
	m, _ := send(pickerWith("alpha", "beta"), key('j'))
	if m.cursor != 1 || m.query.Value() != "" {
		t.Fatalf("browsing: j should move, not type (cursor %d, query %q)", m.cursor, m.query.Value())
	}
}

func TestFilterIsCaseInsensitiveSubstring(t *testing.T) {
	m, _ := send(pickerWith("git status", "https://GitHub.com", "ls -la"), key('/'))
	m = typeText(m, "git")
	got := visibleTexts(m)
	if len(got) != 2 || got[0] != "git status" || got[1] != "https://GitHub.com" {
		t.Fatalf("want [git status https://GitHub.com], got %q", got)
	}
}

func TestEnterAppliesFilterAndReturnsToBrowsing(t *testing.T) {
	m, _ := send(pickerWith("alpha", "jq '.x'", "jest run"), key('/'))
	m = typeText(m, "je")
	m, cmd := send(m, key(tea.KeyEnter))
	if m.filtering {
		t.Fatal("enter while filtering should go back to browsing")
	}
	if m.chosen != "" || cmd != nil {
		t.Fatal("enter while filtering must NOT choose or quit")
	}
	if got := visibleTexts(m); len(got) != 1 || got[0] != "jest run" {
		t.Fatalf("filter should stay applied: got %q", got)
	}
	// Now browsing again: enter chooses from the FILTERED list.
	m, _ = send(m, key(tea.KeyEnter))
	if m.chosen != "jest run" {
		t.Fatalf("want chosen \"jest run\", got %q", m.chosen)
	}
}

func TestEscClearsFilter(t *testing.T) {
	m, _ := send(pickerWith("alpha", "beta"), key('/'))
	m = typeText(m, "be")
	m, cmd := send(m, key(tea.KeyEscape))
	if m.filtering || m.query.Value() != "" {
		t.Fatalf("esc while filtering: want browsing with empty query, got filtering=%v query=%q",
			m.filtering, m.query.Value())
	}
	if cmd != nil {
		t.Fatal("esc while filtering must not quit the picker")
	}
	if len(m.visible()) != 2 {
		t.Fatal("after esc the full list should be back")
	}
}

func TestTypingResetsCursor(t *testing.T) {
	m, _ := send(pickerWith("a1", "a2", "a3"), key('j'), key('j'), key('/'))
	m = typeText(m, "a")
	if m.cursor != 0 {
		t.Fatalf("changing the query should put the cursor back on top, got %d", m.cursor)
	}
}

func TestViewShowsOnlyFilteredEntries(t *testing.T) {
	m, _ := send(pickerWith("alpha", "beta"), key('/'))
	m = typeText(m, "be")
	m.width, m.height = 60, 20 // as if a WindowSizeMsg arrived
	screen := m.View().Content
	if !strings.Contains(screen, "beta") {
		t.Fatalf("filtered entry missing from screen:\n%s", screen)
	}

	if strings.Contains(screen, "alpha") {
		t.Fatalf("screen still shows entries the filter hides (enter would copy something else):\n%s", screen)
	}
}
