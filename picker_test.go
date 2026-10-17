package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// key builds the message Bubble Tea sends for a key press.
func key(code rune) tea.KeyPressMsg {
	if code >= 'a' && code <= 'z' {
		return tea.KeyPressMsg{Code: code, Text: string(code)}
	}
	return tea.KeyPressMsg{Code: code}
}

// send feeds messages through Update, like Bubble Tea's event loop would.
func send(m model, msgs ...tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	for _, msg := range msgs {
		var next tea.Model
		next, cmd = m.Update(msg)
		m = next.(model)
	}
	return m, cmd
}

func threeEntries() model {
	return model{entries: []Entry{{Text: "alpha"}, {Text: "beta"}, {Text: "gamma"}}}
}

func TestCursorMoves(t *testing.T) {
	m, _ := send(threeEntries(), key(tea.KeyDown), key('j'))
	if m.cursor != 2 {
		t.Fatalf("down, j: want cursor 2, got %d", m.cursor)
	}
	m, _ = send(m, key(tea.KeyUp))
	if m.cursor != 1 {
		t.Fatalf("up: want cursor 1, got %d", m.cursor)
	}
}

func TestCursorStaysInBounds(t *testing.T) {
	m, _ := send(threeEntries(), key(tea.KeyUp), key('k'))
	if m.cursor != 0 {
		t.Fatalf("up at top: want 0, got %d", m.cursor)
	}
	m, _ = send(m, key('j'), key('j'), key('j'), key('j'))
	if m.cursor != 2 {
		t.Fatalf("down past bottom: want 2, got %d", m.cursor)
	}
}

func TestEnterChoosesAndQuits(t *testing.T) {
	m, cmd := send(threeEntries(), key('j'), key(tea.KeyEnter))
	if m.chosen != "beta" {
		t.Fatalf("want chosen beta, got %q", m.chosen)
	}
	if cmd == nil {
		t.Fatal("enter should return tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("enter should return tea.Quit")
	}
}

func TestEscQuitsWithoutChoosing(t *testing.T) {
	m, cmd := send(threeEntries(), key('j'), key(tea.KeyEscape))
	if m.chosen != "" {
		t.Fatalf("esc must not choose anything, got %q", m.chosen)
	}
	if cmd == nil {
		t.Fatal("esc should quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("esc should return tea.Quit")
	}
}

func TestEnterOnEmptyHistory(t *testing.T) {
	m, cmd := send(model{}, key(tea.KeyEnter))
	if m.chosen != "" || cmd == nil {
		t.Fatalf("enter on empty: want quit with nothing chosen, got %q", m.chosen)
	}
}

func TestAge(t *testing.T) {
	cases := map[time.Duration]string{
		30 * time.Second: "now",
		12 * time.Minute: "12m",
		7 * time.Hour:    "7h",
		50 * time.Hour:   "2d",
	}
	for d, want := range cases {
		if got := age(d); got != want {
			t.Errorf("age(%v) = %q, want %q", d, got, want)
		}
	}
}

func ctrl(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Mod: tea.ModCtrl}
}

func TestCtrlDigitCopiesRow(t *testing.T) {
	m, cmd := send(threeEntries(), ctrl('2'))
	if m.chosen != "beta" || cmd == nil {
		t.Fatalf("⌃2: want chosen beta and quit, got %q", m.chosen)
	}
}

func TestCtrlZeroIsTenthRow(t *testing.T) {
	m := model{}
	for i := 0; i < 12; i++ {
		m.entries = append(m.entries, Entry{Text: string(rune('a' + i))})
	}
	m, _ = send(m, ctrl('0'))
	if m.chosen != "j" {
		t.Fatalf("⌃0: want tenth row j, got %q", m.chosen)
	}
}

func TestCtrlDigitMissingRowDoesNothing(t *testing.T) {
	m, cmd := send(threeEntries(), ctrl('9'))
	if m.chosen != "" || cmd != nil {
		t.Fatalf("⌃9 on 3 rows: want nothing, got %q", m.chosen)
	}
}

func TestCtrlDigitWhileFiltering(t *testing.T) {
	m := threeEntries()
	m.query.SetValue("a")
	m.filtering = true
	m, _ = send(m, ctrl('2'))
	if m.chosen != "beta" {
		t.Fatalf("⌃2 on filtered [alpha beta gamma]: want beta, got %q", m.chosen)
	}
}

func TestShortcutLabels(t *testing.T) {
	cases := map[int]string{0: "⌃1", 1: "⌃2", 8: "⌃9", 9: "⌃0", 10: ""}
	for idx, want := range cases {
		if got := shortcutLabel(idx); got != want {
			t.Fatalf("row %d: want %q, got %q", idx, want, got)
		}
	}
}
