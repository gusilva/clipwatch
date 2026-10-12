<p align="center">
  <img src="docs/clipboard.png" alt="clipwatch icon" width="128">
</p>

<h1 align="center">clipwatch</h1>

<p align="center">
  Clipboard history for macOS, right inside a <a href="https://herdr.dev">herdr</a> popup.
</p>

<p align="center">
  <img alt="platform" src="https://img.shields.io/badge/platform-macOS-lightgrey?logo=apple">
  <img alt="herdr" src="https://img.shields.io/badge/herdr-%E2%89%A5%200.9.3-blue">
  <img alt="Go" src="https://img.shields.io/badge/Go-cgo-00ADD8?logo=go&logoColor=white">
  <img alt="Swift" src="https://img.shields.io/badge/Swift-swiftc-F05138?logo=swift&logoColor=white">
  <a href="LICENSE"><img alt="license" src="https://img.shields.io/badge/license-MIT-green"></a>
</p>

---

## Features

- **Recent history**: keeps your last 100 copied texts.
- **Quick picker**: open it with one key from any herdr session, then choose an entry to put it back on the clipboard.
- **Privacy aware**: skips anything marked as concealed or transient by the source app, such as password managers ([nspasteboard.org](http://nspasteboard.org) markers).
- **Always on**: the watcher is a background item of `Clipwatch.app`, so it keeps recording after herdr closes. It shows up in Login Items with its own name and icon.
- **Light**: polls the clipboard every 300 ms, and never copies oversized entries into memory.

## Requirements

- macOS 13 or later (the watcher is registered with `SMAppService`)
- herdr ≥ 0.9.3
- Go
- Xcode Command Line Tools (`xcode-select --install`). herdr builds the plugin from source on install. The build needs `swiftc` for the Swift library and `codesign` to sign the app bundle.

## Install

```sh
herdr plugin install gusilva/clipwatch
```

Add a key to `~/.config/herdr/config.toml`:

```toml
[[keys.command]]
key = "prefix+shift+v"
type = "plugin_action"
command = "gusilva.clipwatch.open"
description = "clipboard history"
```

Then reload the config:

```sh
herdr server reload-config
```

The build produces `Clipwatch.app` inside the plugin folder (there is no `bin/` folder). The watcher is registered the next time the herdr server starts. If macOS asks, allow **Clipwatch** in **System Settings → General → Login Items & Extensions**.

## Usage

Press your key (`prefix+shift+v` in the example). The popup lists your history. Pick an entry to put it back on the clipboard.

The binary is `Clipwatch.app/Contents/MacOS/clipwatch`, and it also works on its own:

| Command                  | What it does                                                   |
| ------------------------ | -------------------------------------------------------------- |
| `clipwatch watch`        | Watches the clipboard and records history                      |
| `clipwatch list`         | Prints the saved history                                       |
| `clipwatch copy N`       | Puts entry `N` back on the clipboard                           |
| `clipwatch pick`         | Opens the interactive picker                                   |
| `clipwatch open`         | Opens the picker popup in herdr                                |
| `clipwatch agent`        | Registers the watcher with macOS, or restarts it if registered |
| `clipwatch agent remove` | Stops the watcher and removes it from Login Items              |

`agent` and `agent remove` only work from the binary inside `Clipwatch.app`.

## How it works

The build packs the Go binary into a small app bundle, `Clipwatch.app`, with an icon and an ad-hoc signature. A herdr startup hook runs `clipwatch agent`, which registers the watcher with macOS (`SMAppService`). The watcher job is described by `Clipwatch.app/Contents/Library/LaunchAgents/gusilva.clipwatch.watcher.plist`, and macOS lists it as **Clipwatch** under **System Settings → General → Login Items & Extensions**.

The watcher polls the clipboard and keeps its files in `~/Library/Application Support/clipwatch/`:

| File           | Content                                     |
| -------------- | ------------------------------------------- |
| `history.json` | The saved history                           |
| `watch.log`    | The watcher log (readable only by you)      |

The watcher is started by macOS, not by herdr, so it does not use herdr's state folder.

## Uninstall

```sh
herdr plugin action invoke gusilva.clipwatch.remove-watcher   # stops the watcher, removes it from Login Items
herdr plugin uninstall gusilva.clipwatch
```

Your history stays in `~/Library/Application Support/clipwatch/`. Delete that folder too if you want it gone.

## Development

Building from source, running locally, refreshing the binary herdr uses, and troubleshooting are covered in [docs/instruction.md](docs/instruction.md).

## License

[MIT](LICENSE)
