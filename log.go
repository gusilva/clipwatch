package main

import (
	"log/slog"
	"os"
)

// fatal logs an error with structured attributes and exits with status 1.
func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
