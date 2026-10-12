package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ---------- model: everything the picker knows ----------

type model struct {
	path      string    // history.json
	entries   []Entry   // newest first, as stored
	modTime   time.Time // file mtime at last load (for reload)
	cursor    int       // index of the highlighted entry
	width     int       // terminal size, from WindowSizeMsg
	height    int
	chosen    string          // set on Enter; main writes it to the clipboard AFTER we quit
	filtering bool            // MODE: true while typing in the filter box
	query     textinput.Model // the filter text; applied in both modes
}

// newModel builds a picker with an empty filter box.
func newModel(path string) model {
	q := textinput.New()
	q.Prompt = "/"
	q.Placeholder = "filter"
	return model{path: path, query: q}
}

// visible is the list after the filter: entries containing the query,
// ignoring case. An empty query shows everything.
func (m model) visible() []Entry {
	q := strings.ToLower(m.query.Value())
	if q == "" {
		return m.entries
	}
	var out []Entry
	for _, e := range m.entries {
		if strings.Contains(strings.ToLower(e.Text), q) {
			out = append(out, e)
		}
	}
	return out
}

// tickMsg arrives once per second; we use it to notice file changes.
type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd { return tick() }

// ---------- update: message in, new model out (no side effects) ----------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 1. Messages that mean the same thing in every mode.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tickMsg:
		m = m.reloadIfChanged()
		return m, tick()
	}
	// 2. Everything else depends on the mode.
	if m.filtering {
		return m.updateFiltering(msg)
	}
	return m.updateBrowsing(msg)
}

// updateBrowsing: keys are commands (j moves, / starts filtering, ⏎ copies).
func (m model) updateBrowsing(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	n := len(m.visible())
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < n-1 {
			m.cursor++
		}
	case "/":
		m.filtering = true
		return m, m.query.Focus() // Focus returns a Cmd (cursor blink)
	case "enter":
		if n > 0 {
			m.chosen = m.visible()[m.cursor].Text
		}
		return m, tea.Quit
	case "esc":
		if m.query.Value() != "" { // first esc clears an applied filter
			m.query.Reset()
			m.cursor = 0
			return m, nil
		}
		return m, tea.Quit
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// updateFiltering: keys are TEXT (j types "j"), except a few control keys.
func (m model) updateFiltering(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "enter": // keep the filter, go back to browsing
			m.filtering = false
			m.query.Blur()
			return m, nil
		case "esc": // drop the filter, go back to browsing
			m.filtering = false
			m.query.Blur()
			m.query.Reset()
			m.cursor = 0
			return m, nil
		case "up", "down": // arrows still move: they can't be typed anyway
			return m.updateBrowsing(msg)
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	// Anything else goes to the text input; if the text changed, restart at the top.
	before := m.query.Value()
	var cmd tea.Cmd
	m.query, cmd = m.query.Update(msg)
	if m.query.Value() != before {
		m.cursor = 0
	}
	return m, cmd
}

// reloadIfChanged re-reads history.json when its mtime moved (the watcher saved).
// Same idea as changeCount: poll something cheap, read the content only on change.
func (m model) reloadIfChanged() model {
	info, err := os.Stat(m.path)
	if err != nil || info.ModTime().Equal(m.modTime) {
		return m
	}
	h, err := Load(m.path)
	if err != nil {
		return m
	}

	// 1. Remember WHAT is selected before the list changes under us.
	var selected string
	if vis := m.visible(); m.cursor < len(vis) {
		selected = vis[m.cursor].Text
	}

	// 2. Swap in the new list.
	m.entries, m.modTime = h.Entries, info.ModTime()

	// 3. Follow that text to its new position.
	for i, e := range m.visible() {
		if e.Text == selected {
			m.cursor = i
			return m
		}
	}

	// 4. It's gone (or nothing was selected): stay put, but inside the list.
	m.cursor = min(m.cursor, max(len(m.visible())-1, 0))
	return m
}

// ---------- view: model in, screen out ----------

var (
	accent   = lipgloss.Color("#89b4fa")
	dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("#7f849c"))
	bold     = lipgloss.NewStyle().Bold(true)
	selected = lipgloss.NewStyle().Background(lipgloss.Color("#313244"))
	dot      = lipgloss.NewStyle().Foreground(accent).Render("●")
)

const rowHeight = 2 // each entry = preview line + meta line

func (m model) View() tea.View {
	w := max(m.width, 40)
	var b strings.Builder

	// header
	b.WriteString(spread(bold.Render("clipboard"), dim.Render(fmt.Sprintf("%d entries", len(m.entries))), w) + "\n")
	b.WriteString(m.filterLine() + "\n")

	// body: only the rows that fit (header 2 lines + footer 2 lines)
	vis := m.visible()
	perPage := max((m.height-4)/rowHeight, 1)
	start := max(0, m.cursor-perPage+1) // scroll so the cursor stays on screen
	end := min(len(vis), start+perPage)
	if len(vis) == 0 {
		b.WriteString(dim.Render("nothing yet — copy something") + "\n")
	}
	for i := start; i < end; i++ {
		e := vis[i]
		mark := " "
		if i == m.cursor {
			mark = dot
		}
		line1 := spread(mark+" "+preview(e.Text, w-12), dim.Render(age(time.Since(e.CopiedAt))), w)
		line2 := "  " + dim.Render(meta(e.Text))
		row := line1 + "\n" + padRight(line2, w)
		if i == m.cursor {
			row = selected.Render(row)
		}
		b.WriteString(row + "\n")
	}

	// footer
	hint := "↑/k ↓/j move · ⏎ copy · esc close"
	if m.filtering {
		hint = "type to filter · ↑↓ move · ⏎ apply · esc clear"
	}
	b.WriteString("\n" + dim.Render(hint))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// filterLine shows the input while typing, the applied query after,
// and nothing when there is no filter.
func (m model) filterLine() string {
	switch {
	case m.filtering:
		return m.query.View()
	case m.query.Value() != "":
		return dim.Render("/" + m.query.Value() + fmt.Sprintf("  · %d of %d", len(m.visible()), len(m.entries)))
	default:
		return ""
	}
}

// spread puts left at the start and right at the end of a line of width w.
func spread(left, right string, w int) string {
	gap := max(w-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

func padRight(s string, w int) string {
	return s + strings.Repeat(" ", max(w-lipgloss.Width(s), 0))
}

// meta describes an entry for the second row: size and line count.
func meta(s string) string {
	lines := strings.Count(s, "\n") + 1
	if lines == 1 {
		return fmt.Sprintf("%d chars", len([]rune(s)))
	}
	return fmt.Sprintf("%d chars · %d lines", len([]rune(s)), lines)
}

// age formats how long ago something happened, like the screenshot: 12m, 2h.
func age(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// ---------- entry point: run the UI, then do the side effect ----------

func pick() {
	path, err := historyPath()
	if err != nil {
		fatal("error", "err", err)
	}
	m := newModel(path).reloadIfChanged() // first load

	final, err := tea.NewProgram(m).Run()
	if err != nil {
		fatal("error", "err", err)
	}
	if text := final.(model).chosen; text != "" {
		if _, err := writeString(text); err != nil {
			fatal("error", "err", err)
		}
	}
}
