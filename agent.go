package main

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// agentJob describes the launchd job that runs the watcher for this plugin.
// Every field is derived from HERDR_PLUGIN_ID, so a fork with another plugin id
// gets its own job.
type agentJob struct {
	label     string // gusilva.clipwatch.watcher
	plistPath string // ~/Library/LaunchAgents/<label>.plist
	domain    string // gui/501: your logged-in GUI session
	target    string // gui/501/<label>
}

// job builds the agentJob for the current plugin. It exits if HERDR_PLUGIN_ID
// is not set, which means the process was not started by herdr.
func job() agentJob {
	id := os.Getenv("HERDR_PLUGIN_ID")
	if id == "" {
		fatal("agent must be run by herdr (HERDR_PLUGIN_ID is not set)")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fatal("error", "err", err)
	}

	label := id + ".watcher"
	domain := fmt.Sprintf("gui/%d", os.Getuid())

	return agentJob{
		label:     label,
		plistPath: filepath.Join(home, "Library", "LaunchAgents", label+".plist"),
		domain:    domain,
		target:    domain + "/" + label,
	}
}

// loaded reports whether launchd currently knows the job.
func (j agentJob) loaded() bool {
	return exec.Command("launchctl", "print", j.target).Run() == nil
}

// unload stops the job and waits until launchd has forgotten it, because
// bootout can return before the job is gone.
func (j agentJob) unload() {
	if !j.loaded() {
		return
	}

	launchctl("bootout", j.target)

	for i := 0; i < 50 && j.loaded(); i++ {
		time.Sleep(100 * time.Millisecond)
	}
}

// installAgent runs as herdr's [[startup]] hook. It hands the watcher to
// launchd, the macOS process supervisor, and then returns, because herdr hooks
// are one-shot.
func installAgent() {
	j := job()

	state := os.Getenv("HERDR_PLUGIN_STATE_DIR")
	if state == "" {
		fatal("agent must be run by herdr (HERDR_PLUGIN_STATE_DIR is not set)")
	}

	exe, err := os.Executable() // absolute path of THIS binary
	if err != nil {
		fatal("error", "err", err)
	}

	want := []byte(plist(j.label, exe, state))
	have, _ := os.ReadFile(j.plistPath) // missing file → nil, which never matches

	if j.loaded() && bytes.Equal(have, want) {
		// Same job description: just restart, so a rebuilt binary takes over.
		launchctl("kickstart", "-k", j.target)
		slog.Info("watcher restarted", "target", j.target)

		return
	}

	// New or changed description (e.g. the binary moved from your linked
	// folder to herdr's install folder). kickstart would keep the OLD one,
	// so unload, rewrite, load.
	j.unload()
	if err := os.MkdirAll(filepath.Dir(j.plistPath), 0o755); err != nil {
		fatal("error", "err", err)
	}

	if err := os.WriteFile(j.plistPath, want, 0o644); err != nil {
		fatal("error", "err", err)
	}

	launchctl("bootstrap", j.domain, j.plistPath)
	slog.Info("watcher is run by launchd", "target", j.target)
}

// removeAgent stops the launchd job and deletes its plist. It is the
// clean-uninstall step.
func removeAgent() {
	j := job()
	j.unload()
	if err := os.Remove(j.plistPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fatal("error", "err", err)
	}

	slog.Info("watcher removed", "target", j.target)
}

// launchctl runs launchctl with args and exits if the command fails.
func launchctl(args ...string) {
	out, err := exec.Command("launchctl", args...).CombinedOutput()
	if err != nil {
		fatal("launchctl failed", "args", strings.Join(args, " "), "err", err, "output", string(out))
	}
}

// plistSource is the launchd job description template, embedded from agent.plist.
// Its KeepAlive/PathState key tells launchd to restart the watcher whenever it
// stops, as long as the binary still exists. A deleted plugin folder therefore
// does not leave launchd retrying every 10 seconds.
//
//go:embed agent.plist
var plistSource string

// plistTemplate is the parsed plistSource. Its xml function escapes values for
// use inside XML elements.
var plistTemplate = template.Must(template.New("agent.plist").Funcs(template.FuncMap{
	"xml": esc,
}).Parse(plistSource))

// plist renders the launchd job description for the watcher. label names the
// job, exe is the path of this binary, and state is the plugin state directory
// that also receives watch.log.
func plist(label, exe, state string) string {
	var b strings.Builder
	err := plistTemplate.Execute(&b, struct{ Label, Exe, State, LogPath string }{
		Label:   label,
		Exe:     exe,
		State:   state,
		LogPath: filepath.Join(state, "watch.log"),
	})
	if err != nil {
		fatal("render plist", "err", err)
	}

	return b.String()
}

// esc returns s escaped for use inside an XML element.
func esc(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))

	return b.String()
}
