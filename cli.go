package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// main runs the subcommand named by the first argument, defaulting to "watch".
func main() {
	cmd := "watch"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	switch cmd {
	case "watch":
		watch()
	case "list":
		list()
	case "copy":
		if len(os.Args) != 3 {
			usage()
		}
		n, err := strconv.Atoi(os.Args[2])
		if err != nil {
			usage()
		}
		copyEntry(n)
	case "pick":
		pick()
	case "open":
		openPopup()
	case "agent":
		if len(os.Args) == 3 && os.Args[2] == "remove" {
			removeAgent()
		}

	default:
		usage()
	}
}

// usage reports an invalid command line and exits with status 2.
func usage() {
	slog.Error("invalid command", "usage", "clipwatch [watch | list | copy N | pick | open | agent [remove]]")
	os.Exit(2)
}

// loadHistory reads the history file (read-only: only `watch` writes it).
func loadHistory() *History {
	path, err := historyPath()
	if err != nil {
		fatal("error", "err", err)
	}
	h, err := Load(path)
	if err != nil {
		fatal("error", "err", err)
	}
	return h
}

func list() {
	for i, e := range loadHistory().Entries {
		fmt.Printf("%3d  %s\n", i+1, preview(e.Text, 60))
	}
}

// copyEntry puts entry n (1-based, as shown by list) on the clipboard.
func copyEntry(n int) {
	h := loadHistory()
	if n < 1 || n > len(h.Entries) {
		fatal("no such entry", "n", n, "entries", len(h.Entries))
	}
	if _, err := writeString(h.Entries[n-1].Text); err != nil {
		fatal("error", "err", err)
	}
	fmt.Printf("copied #%d: %s\n", n, preview(h.Entries[n-1].Text, 60))
}

// preview squeezes text onto one line of at most max characters (runes).
// Runs of whitespace, including newlines, become one space. Text that is too
// long is cut and ends with "…".
func preview(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= max {
		return s
	}

	r := []rune(s)
	return string(r[:max-1]) + "…"
}
