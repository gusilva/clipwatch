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

// model is the whole state of the picker: the history entries, the cursor, the
// filter, and whether the watcher is alive. Its path is history.json and its
// heartbeat is the watcher's heartbeat file. The entries are newest first, as
// stored, and modTime is the mtime of history.json at the last load. The width
// and height come from the terminal size. The chosen text is set on Enter, and
// pick writes it to the clipboard after the UI has quit. While filtering is
// true, keys are typed into the query, which filters the list in both modes.
type model struct {
	path      string
	heartbeat string
	entries   []Entry
	modTime   time.Time
	cursor    int
	width     int
	height    int
	chosen    string
	filtering bool
	query     textinput.Model
	watching  bool
}

// newModel builds a picker with an empty filter box.
func newModel(path string) model {
	q := textinput.New()
	q.Prompt = "/"
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

// tickMsg arrives once per second, to notice file changes and the heartbeat.
type tickMsg time.Time

// tick schedules the next tickMsg.
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Init starts the tick loop.
func (m model) Init() tea.Cmd { return tick() }

// Update handles window size and ticks in every mode, and sends everything
// else to the handler of the current mode.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tickMsg:
		m = m.reloadIfChanged()
		m.watching = watcherAlive(m.heartbeat, time.Now())
		return m, tick()
	}

	if k, ok := msg.(tea.KeyPressMsg); ok {
		if i, ok := shortcutIndex(k); ok {
			return m.choose(i)
		}
	}

	if m.filtering {
		return m.updateFiltering(msg)
	}

	return m.updateBrowsing(msg)
}

// shortcutRows is how many rows have a shortcut: ctrl+1 to ctrl+9, then ctrl+0.
const shortcutRows = 10

// shortcutIndex maps ctrl+1…ctrl+9 to rows 0…8 and ctrl+0 to row 9. It works in both
// modes, because a key with ctrl+digit cannot be typed as text.
func shortcutIndex(k tea.KeyPressMsg) (int, bool) {
	if k.Mod&tea.ModCtrl == 0 || k.Code < '0' || k.Code > '9' {
		return 0, false
	}

	return (int(k.Code-'0') + shortcutRows - 1) % shortcutRows, true
}

// shortcutLabel is the label of row i, like "⌃1", and "" for rows without one.
func shortcutLabel(i int) string {
	if i < 0 || i >= shortcutRows {
		return ""
	}

	return fmt.Sprintf("⌃%d", (i+1)%shortcutRows)
}

// choose copies row i of the visible list and quits. It does nothing when the
// row does not exist.
func (m model) choose(i int) (tea.Model, tea.Cmd) {
	vis := m.visible()
	if i >= len(vis) {
		return m, nil
	}

	m.cursor, m.chosen = i, vis[i].Text
	return m, tea.Quit
}

// updateBrowsing handles keys as commands: j moves, / starts filtering, and
// Enter copies. The first Esc clears an applied filter, and the next one quits.
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
		return m, m.query.Focus()
	case "enter":
		if n > 0 {
			m.chosen = m.visible()[m.cursor].Text
		}

		return m, tea.Quit
	case "esc":
		if m.query.Value() != "" {
			return m.clearFilter(), nil
		}

		return m, tea.Quit
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// updateFiltering handles keys as text: j types "j". Enter keeps the filter
// and Esc drops it, and both go back to browsing. The arrows still move,
// because they cannot be typed. A change of the text sends the cursor to the top.
func (m model) updateFiltering(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "enter":
			m.filtering = false
			m.query.Blur()
			return m, nil
		case "esc":
			m.filtering = false
			m.query.Blur()
			return m.clearFilter(), nil
		case "up", "down":
			return m.updateBrowsing(msg)
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	before := m.query.Value()
	var cmd tea.Cmd
	m.query, cmd = m.query.Update(msg)
	if m.query.Value() != before {
		m.cursor = 0
	}

	return m, cmd
}

// clearFilter empties the filter text and puts the cursor on the first entry.
func (m model) clearFilter() model {
	m.query.Reset()
	m.cursor = 0
	return m
}

// reloadIfChanged re-reads history.json when its mtime moved, which means the
// watcher saved. It polls something cheap and reads the content only on change.
// The cursor follows the selected text to its new position.
func (m model) reloadIfChanged() model {
	info, err := os.Stat(m.path)
	if err != nil || info.ModTime().Equal(m.modTime) {
		return m
	}

	h, err := Load(m.path)
	if err != nil {
		return m
	}

	selected := m.selectedText()
	m.entries, m.modTime = h.Entries, info.ModTime()
	return m.followText(selected)
}

// selectedText returns the text under the cursor, or "" when the list is empty.
func (m model) selectedText() string {
	vis := m.visible()
	if m.cursor >= len(vis) {
		return ""
	}

	return vis[m.cursor].Text
}

// followText moves the cursor to the entry with the given text. When that
// entry is gone, or nothing was selected, the cursor stays where it was but
// inside the list.
func (m model) followText(text string) model {
	for i, e := range m.visible() {
		if e.Text == text {
			m.cursor = i
			return m
		}
	}

	m.cursor = min(m.cursor, max(len(m.visible())-1, 0))
	return m
}

// Colours of the picker. pageBg is the popup background, the same as the herdr
// frame, and selBg is the background of the highlighted entry.
var (
	accent     = lipgloss.Color("#89b4fa")
	dim        = lipgloss.NewStyle().Foreground(lipgloss.Color("#7f849c"))
	pageBg     = lipgloss.Color("#191824")
	selBg      = lipgloss.Color("#313244")
	selected   = lipgloss.NewStyle().Background(selBg)
	labelStyle = lipgloss.NewStyle().Foreground(accent)
	ruleText   = dim.Foreground(lipgloss.Color("#45475a"))
)

// rowHeight is the lines per entry: the preview line, the meta line and the
// separator rule.
const rowHeight = 3

// chromeHeight is the lines around the list: two for the header with its blank
// line, and two for the footer with its blank line.
const chromeHeight = 4

// hint lists the keys in browsing mode.
const hint = "↑/k ↓/j move · / search · ⏎ copy · ⌃1-0 copy row · esc close"

// View draws the header, the visible rows, and the footer, which is pinned to
// the last row of the screen. Only the rows that fit are drawn, and the list
// scrolls so that the cursor stays on screen.
func (m model) View() tea.View {
	w := max(m.width, 40)
	var b strings.Builder

	count := dim.Render(fmt.Sprintf("%d entries", len(m.entries)))
	b.WriteString(spread(m.filterLine(), count, w) + "\n\n")
	b.WriteString(m.rows(w))
	b.WriteString(m.footer(strings.Count(b.String(), "\n"), w))

	v := tea.NewView(paintBackground(b.String(), w, max(m.height, 1)))
	v.AltScreen = true
	v.BackgroundColor = pageBg
	return v
}

// rows renders the entries that fit on the screen, one rowHeight block each,
// with a rule between entries and none after the last. It renders a hint when
// the list is empty.
func (m model) rows(w int) string {
	vis := m.visible()
	if len(vis) == 0 {
		return dim.Render("nothing yet — copy something") + "\n"
	}

	perPage := max((m.height-chromeHeight)/rowHeight, 1)
	start := max(0, m.cursor-perPage+1)
	end := min(len(vis), start+perPage)

	var b strings.Builder
	for i := start; i < end; i++ {
		b.WriteString(renderRow(vis[i], shortcutLabel(i), i == m.cursor, w) + "\n")
		if i < end-1 {
			b.WriteString(ruleText.Render(strings.Repeat("─", w)) + "\n")
		}
	}

	return b.String()
}

// footer renders the key hints with the watcher indicator at the right. The
// used lines already filled by the header and the rows are padded with blank
// lines, at least one, so that the footer lands on the last row of the screen.
func (m model) footer(used, w int) string {
	gap := max(m.height-used-1, 1)
	return strings.Repeat("\n", gap) + spread(dim.Render(hint), watchIndicator(m.watching), w)
}

// renderRow renders one entry as a preview line with its age, and a meta line.
// The label ("⌃1") takes the place of the marker; rows without one get blanks.
// A highlighted entry gets the selection background across the whole row. Every
// piece carries that background itself, because a nested style ends with a
// reset, which would cancel an outer background.
func renderRow(e Entry, label string, highlighted bool, w int) string {
	fill, dimSt := lipgloss.NewStyle(), dim
	if highlighted {
		fill, dimSt = selected, dim.Background(selBg)
	}

	mark := fill.Render("  ")
	if label != "" {
		mark = labelStyle.Background(fill.GetBackground()).Render(label)
	}

	left := fill.Render(" ") + mark + fill.Render(" "+preview(e.Text, w-15))
	right := dimSt.Render(age(time.Since(e.CopiedAt)))
	gap := fill.Render(strings.Repeat(" ", max(w-lipgloss.Width(left)-lipgloss.Width(right), 1)))

	metaLine := fill.Render("  ") + dimSt.Render(meta(e.Text))
	metaLine += fill.Render(strings.Repeat(" ", max(w-lipgloss.Width(metaLine), 0)))

	return left + gap + right + "\n" + metaLine
}

// paintBackground gives every cell of a w x h screen the page background.
// Styled pieces end with a reset, which would drop the background, so it
// switches the background back on after each reset and pads every line to the
// full width.
func paintBackground(s string, w, h int) string {
	on := backgroundOn()
	lines := strings.Split(s, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}

	for i, l := range lines {
		l = strings.ReplaceAll(l, "\x1b[m", "\x1b[m"+on)
		l = strings.ReplaceAll(l, "\x1b[0m", "\x1b[0m"+on)
		lines[i] = on + l + strings.Repeat(" ", max(w-lipgloss.Width(l), 0)) + "\x1b[m"
	}

	return strings.Join(lines, "\n")
}

// backgroundOn returns the escape sequence that turns the page background on.
// It renders a space with that background and keeps what comes before the space.
func backgroundOn() string {
	on, _, _ := strings.Cut(lipgloss.NewStyle().Background(pageBg).Render(" "), " ")
	return on
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

// watchIndicator is the footer badge at the bottom right: a green circle when
// the watcher runs, and a red circle when not. It is only the emoji, so it fits
// next to the key hints in a narrow popup.
func watchIndicator(watching bool) string {
	if watching {
		return "\U0001F7E2"
	}

	return "\U0001F534"
}

// spread puts left at the start and right at the end of a line of width w.
func spread(left, right string, w int) string {
	gap := max(w-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
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

// pick runs the picker UI, and then, if an entry was chosen, writes it to the
// clipboard. The write happens after the UI has quit.
func pick() {
	path, err := historyPath()
	if err != nil {
		fatal("error", "err", err)
	}

	beat, err := heartbeatPath()
	if err != nil {
		fatal("error", "err", err)
	}

	m := newModel(path).reloadIfChanged()
	m.heartbeat = beat
	m.watching = watcherAlive(beat, time.Now())

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
