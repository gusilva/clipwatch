# Development instructions

How to build and run clipwatch locally, make herdr pick up your latest changes, and fix common problems.

## 1. Build and launch locally

Requirements: Go and the Xcode Command Line Tools (`xcode-select --install`).

```sh
cd ~/clipwatch
./build.sh          # builds the Swift shim (build/libpasteboard.a), then bin/clipwatch
go test ./...
```

Plain `go build` fails until `./build.sh` has run once, because cgo links `build/libpasteboard.a`.

Try the pieces without herdr:

```sh
./bin/clipwatch watch      # records the clipboard in the foreground (Ctrl+C to stop)
./bin/clipwatch list       # prints the saved history
./bin/clipwatch copy 1     # puts entry 1 back on the clipboard
./bin/clipwatch pick       # opens the picker in this terminal
```

Link the folder into herdr, so it runs your build:

```sh
herdr plugin link .
herdr server reload-config
```

## 2. Which clipwatch is herdr running?

```sh
herdr plugin list
```

- `[local:/Users/…/clipwatch]`: herdr runs your folder. Rebuilding is enough (section 3).
- `[github:gusilva/clipwatch@…]`: herdr runs its own installed copy, and rebuilding your folder changes nothing. See section 4.

## 3. Refresh the binary after a change (linked folder)

Always rebuild first:

```sh
./build.sh && go test ./...
```

Then, depending on what you changed:

| You changed…                                                  | Extra step                                                      | Why                                                                                |
| ------------------------------------------------------------- | --------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `picker.go`, `herdr.go` (`pick`, `open`)                      | None. Open the popup again                                      | Every popup starts a new `bin/clipwatch` process                                   |
| `main.go`, `history.go`, `pasteboard.go`, `swift/*` (watcher) | `launchctl kickstart -k gui/$(id -u)/gusilva.clipwatch.watcher` | The running watcher is still the old process. This restarts it with the new binary |
| `agent.go` (the plist)                                        | Run the hook by hand (below), or restart the herdr server       | The job file has to be reloaded, and `kickstart` doesn't re-read it                |
| `herdr-plugin.toml`, `build.sh`                               | `herdr plugin unlink gusilva.clipwatch && herdr plugin link .`  | herdr reads the manifest when you link the plugin                                  |

Run the startup hook by hand:

```sh
HERDR_PLUGIN_ID=gusilva.clipwatch \
HERDR_PLUGIN_STATE_DIR="$HOME/.local/state/herdr/plugins/gusilva.clipwatch" \
./bin/clipwatch agent
```

Check that the new watcher is running:

```sh
launchctl print gui/$(id -u)/gusilva.clipwatch.watcher | grep -E "^\s(state|program|runs)"
tail -f ~/.local/state/herdr/plugins/gusilva.clipwatch/watch.log   # then copy something
```

`runs` goes up by one after a kickstart. `program` should point at your folder.

## 4. Refresh when installed from GitHub

Pick one:

- **Ship the change:** push to `main`, then:
  ```sh
  herdr plugin uninstall gusilva.clipwatch && herdr plugin install gusilva/clipwatch
  ```
  Restart the herdr server. There is no `plugin update` in v1.
- **Keep developing (simpler):** `herdr plugin uninstall gusilva.clipwatch`, then `herdr plugin link .`, then restart the server. The startup hook moves the watcher back to your folder, and you use section 3 from then on.

## 5. Troubleshooting

**`swiftc: command not found`, or `xcrun: error` during the build**
Install the tools with `xcode-select --install`. If they are installed but not selected, run `sudo xcode-select --switch /Library/Developer/CommandLineTools`.

**Link errors such as `library 'pasteboard' not found` or `undefined: _pbChangeCount`**
`build/libpasteboard.a` is missing or stale. Run `./build.sh` (not `go build`). If it persists, `rm -rf build bin && ./build.sh`.

**Swift or SDK version mismatch errors**
After a macOS or Xcode update, remove the old build output and rebuild: `rm -rf build bin && ./build.sh`.

**Popup is empty or history is not growing**

1. Check that the watcher is loaded:
   ```sh
   launchctl print gui/$(id -u)/gusilva.clipwatch.watcher | grep -E "^\s(state|program|runs)"
   ```
2. Read its log: `tail -f ~/.local/state/herdr/plugins/gusilva.clipwatch/watch.log`.
3. Restart it: `launchctl kickstart -k gui/$(id -u)/gusilva.clipwatch.watcher`.
4. If it is not loaded, run the startup hook by hand (section 3), or restart the herdr server.

Copies from password managers are skipped on purpose. Entries over the size limit are skipped too.

**My change has no effect**
You are probably running another copy. Check `herdr plugin list` (section 2). For the watcher, kickstart it (section 3).

**The key does nothing**
Check the `[[keys.command]]` block in `~/.config/herdr/config.toml`, and run `herdr server reload-config`. The command id is `gusilva.clipwatch.open`.

**Watcher still runs after uninstalling**

```sh
herdr plugin action invoke gusilva.clipwatch.remove-watcher
```

If herdr is already gone, remove it manually:

```sh
launchctl bootout gui/$(id -u)/gusilva.clipwatch.watcher
rm ~/Library/LaunchAgents/gusilva.clipwatch.watcher.plist
```

**Reset the history**
Delete `~/.local/state/herdr/plugins/gusilva.clipwatch/` (stop the watcher first, or it recreates the file).
