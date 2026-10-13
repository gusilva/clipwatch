# Development instructions

How to build and run clipwatch locally, make herdr pick up your latest changes, and fix common problems.

## 1. How the build is laid out

`./build.sh` produces `Clipwatch.app`, a small macOS app bundle.

```
Clipwatch.app/
└── Contents/
    ├── MacOS/clipwatch                  the Go binary (every command runs from here)
    ├── Info.plist                       app name, id and icon (from app/Info.plist)
    ├── Library/LaunchAgents/
    │   └── gusilva.clipwatch.watcher.plist   the watcher job (from app/)
    └── Resources/Clipwatch.icns         the icon shown in Login Items
```

The bundle is needed because macOS registers the watcher with `SMAppService`, and that only works for a binary inside an app. `Clipwatch.app/` and `build/` are build output and are in `.gitignore`.

Runtime files are in `~/Library/Application Support/clipwatch/`: `history.json`, `watch.log` and `heartbeat`. The watcher touches `heartbeat` every second, and the picker shows 🟢 when it is less than 3 seconds old and 🔴 otherwise. It does not depend on launchd or its output format.

## 2. Build and launch locally

Requirements: Go and the Xcode Command Line Tools (`xcode-select --install`). The build uses `swiftc` and `codesign`. macOS 13 or later.

```sh
cd ~/Documents/teaching/herdr-plugin/clipwatch
./build.sh          # Swift library (build/libpasteboard.a) → Go binary → Clipwatch.app, signed
go test ./...
```

Plain `go build` and `go test` fail until `./build.sh` has run once, because cgo links `build/libpasteboard.a`.

Try the pieces without herdr:

```sh
CW=./Clipwatch.app/Contents/MacOS/clipwatch
$CW watch      # records the clipboard in the foreground (Ctrl+C to stop)
$CW list       # prints the saved history
$CW copy 1     # puts entry 1 back on the clipboard
$CW pick       # opens the picker in this terminal
$CW agent      # registers the watcher with macOS, or restarts it
```

Always run `agent` and `agent remove` from the binary inside `Clipwatch.app`. A binary built with plain `go build`, or started with `go run`, fails with `register failed`.

Link the folder into herdr, so it runs your build:

```sh
herdr plugin link .
herdr server reload-config
```

## 3. Which clipwatch is herdr running?

```sh
herdr plugin list
```

- `[local:/Users/…/clipwatch]`: herdr runs your folder. Rebuilding is enough (section 4).
- `[github:gusilva/clipwatch@…]`: herdr runs its own installed copy, and rebuilding your folder changes nothing. See section 5.

## 4. Refresh the binary after a change (linked folder)

Always rebuild first:

```sh
./build.sh && go test ./...
```

Then, depending on what you changed:

| You changed…                                                                               | Extra step                                                     | Why                                                                                   |
| ------------------------------------------------------------------------------------------ | -------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| `picker.go`, `herdr.go` (`pick`, `open`)                                                   | None. Open the popup again                                     | Every popup starts a new `clipwatch` process                                          |
| `main.go`, `history.go`, `datadir.go`, `pasteboard.go`, `swift/pasteboard.swift` (watcher) | `./Clipwatch.app/Contents/MacOS/clipwatch agent`               | The running watcher is still the old process. `agent` restarts it with the new binary |
| `agent.go`, `service.go`, `swift/service.swift`, `app/*.plist`                             | `clipwatch agent remove`, then `clipwatch agent` (same binary) | macOS reads the job description when the watcher is registered, so register it again  |
| `herdr-plugin.toml`, `build.sh`                                                            | `herdr plugin unlink gusilva.clipwatch && herdr plugin link .` | herdr reads the manifest when you link the plugin                                     |

The `agent` command is also what herdr runs at server start, so restarting the herdr server has the same effect.

Check that the new watcher is running:

```sh
launchctl print gui/$(id -u)/gusilva.clipwatch.watcher | grep -E "^\s(state|program|runs)"
tail -f ~/Library/Application\ Support/clipwatch/watch.log   # then copy something
```

`state` should be `running`, `program identifier` should be `Contents/MacOS/clipwatch`, and `runs` goes up by one after each restart.

## 5. Refresh when installed from GitHub

Pick one:

- **Ship the change:** push to `main`, then:
  ```sh
  herdr plugin uninstall gusilva.clipwatch && herdr plugin install gusilva/clipwatch
  ```
  Restart the herdr server. There is no `plugin update` in v1.
- **Keep developing (simpler):** `herdr plugin uninstall gusilva.clipwatch`, then `herdr plugin link .`, then restart the server. The startup hook registers the watcher from your folder, and you use section 4 from then on.

## 6. Versioning and releases

`version` in `herdr-plugin.toml` is the only place where the version is written. `./build.sh` copies it into `Clipwatch.app/Contents/Info.plist` as `CFBundleShortVersionString`, and sets `CFBundleVersion` to the number of commits in the git history (`1` if git history is not available). `app/Info.plist` holds the placeholder `0.0.0`, so do not edit versions there.

Changes are listed in `CHANGELOG.md`, newest first, with a heading `## X.Y.Z - YYYY-MM-DD` per version and the sections Added, Changed, Fixed or Internal. `scripts/release.sh` does not touch it, so write the entry for the new version by hand and commit it **before** you run the script.

A git tag cannot change files, so the version is bumped first and the tag is made from that commit. `scripts/release.sh` does both:

```sh
# 1. add "## 0.3.0 - 2026-10-12" to CHANGELOG.md and commit it
scripts/release.sh 0.3.0        # sets version, commits "chore(release): v0.3.0", tags v0.3.0
git push && git push origin v0.3.0
```

The script stops if the working tree is not clean or the tag already exists. It does not push.

Check the installed version:

```sh
/usr/libexec/PlistBuddy -c "Print :CFBundleShortVersionString" Clipwatch.app/Contents/Info.plist
```

## 7. Checks

```sh
scripts/check.sh
```

Runs `sh -n` on the scripts, `golangci-lint`, `swiftlint`, a check for stale paths in the docs, `./build.sh` and `go test`. A pre-commit hook and the GitHub Actions workflow in `.github/workflows/check.yml` run it too.

After cloning, enable the hook once:

```sh
git config core.hooksPath .githooks
```

Needs `golangci-lint` and `swiftlint` (`brew install golangci-lint swiftlint`).

## 8. Troubleshooting

**`swiftc: command not found`, `codesign: command not found`, or `xcrun: error` during the build**
Install the tools with `xcode-select --install`. If they are installed but not selected, run `sudo xcode-select --switch /Library/Developer/CommandLineTools`.

**Link errors such as `library 'pasteboard' not found` or `undefined: _pbChangeCount`**
`build/libpasteboard.a` is missing or stale. Run `./build.sh` (not `go build`). If it persists, `rm -rf build Clipwatch.app && ./build.sh`.

**Swift or SDK version mismatch errors**
After a macOS or Xcode update, remove the old build output and rebuild: `rm -rf build Clipwatch.app && ./build.sh`.

**`register failed (is this binary inside Clipwatch.app?)`**
`agent` ran from a binary outside the bundle. Use `./Clipwatch.app/Contents/MacOS/clipwatch agent`, and rebuild with `./build.sh` if the bundle is missing.

**`watcher is switched off: enable Clipwatch in System Settings…`**
You (or macOS) turned the Clipwatch background item off. Switch it on under **System Settings → General → Login Items & Extensions**, then run `agent` again. Clipwatch opens that page for you when it detects this.

**Clipwatch is missing from Login Items, or has no name or icon**
Check that `Clipwatch.app/Contents/Info.plist` and `Resources/Clipwatch.icns` exist (rebuild if not). Then run `clipwatch agent remove` and `clipwatch agent` to register it again.

**The 🔴 in the picker, or `state = spawn scheduled` and a huge `runs` count**
The watcher is registered but macOS refuses to start it. `launchctl print` shows `last exit code = 78: EX_CONFIG` or `last exit reason = OS_REASON_CODESIGNING`, and `watch.log` stops growing. Running `clipwatch watch` by hand works. To see the reason:

```sh
log show --last 2m --predicate 'process == "launchd"' | grep -i clipwatch
```

A line with `OS_REASON_CODESIGNING | Launch Constraint Violation` is this problem. It came back after each `./build.sh` or `scripts/check.sh`, which sign the bundle again. Restarting with `agent` was not enough. Registering again fixed it:

```sh
CW=./Clipwatch.app/Contents/MacOS/clipwatch
$CW agent remove
$CW agent
```

Then check that `state = running`. The first time, `xattr -cr Clipwatch.app` and `codesign --force --sign - Clipwatch.app` were run between the two commands. The second time, those two alone did not help, but `agent remove` followed by `agent` did. Why macOS refuses the rebuilt bundle was not found.

**Popup is empty or history is not growing**

1. Check that the watcher is loaded:
   ```sh
   launchctl print gui/$(id -u)/gusilva.clipwatch.watcher | grep -E "^\s(state|program|runs)"
   ```
2. Read its log: `tail -f ~/Library/Application\ Support/clipwatch/watch.log`.
3. Restart it: `./Clipwatch.app/Contents/MacOS/clipwatch agent`.
4. If it is not loaded, run `agent` by hand (section 2), or restart the herdr server.

Copies from password managers are skipped on purpose. Entries over the size limit are skipped too.

**My change has no effect**
You are probably running another copy. Check `herdr plugin list` (section 3). For the watcher, restart it with `agent` (section 4).

**The key does nothing**
Check the `[[keys.command]]` block in `~/.config/herdr/config.toml`, and run `herdr server reload-config`. The command id is `gusilva.clipwatch.open`.

**Watcher still runs after uninstalling**

```sh
herdr plugin action invoke gusilva.clipwatch.remove-watcher
```

If herdr is already gone, run `agent remove` from the binary in your folder, or switch Clipwatch off under **System Settings → General → Login Items & Extensions**.

**Reset the history**
Delete `~/Library/Application Support/clipwatch/history.json`. Stop the watcher first (`agent remove`), or it keeps its copy in memory and writes it back on the next clip.
