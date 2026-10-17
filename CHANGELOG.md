# Changelog

All notable changes to clipwatch. The version lives in `herdr-plugin.toml`.

## v0.5.0 - 2026-10-17

### Added

- Picker: `ctrl+1` to `ctrl+9` copy rows 1 to 9, and `ctrl+0` copies row 10.
  The key follows the row's place in the list, so `ctrl+2` always copies the
  second row. It works while the filter box is open, and it does nothing when
  the row does not exist. Ghostty and herdr do not pass `⌘+digit` to the
  program, so the shortcut uses Ctrl.
- Picker: each of the first ten rows shows its key (`⌃1` to `⌃0`) in place of
  the selection dot. The footer hint lists `⌃1-0 copy row`.

### Changed

- Picker: the selected row is marked only by its background, because the dot
  is gone.
- Plugin: the picker popup is 40% wide instead of 20%.

## v0.4.0 - 2026-10-13

### Added

- Picker: the bottom right corner shows 🟢 when the watcher is running
  and 🔴 when it is not. The watcher touches a `heartbeat` file in
  its data folder every second, and the picker shows green while the file is
  less than 3 seconds old. It does not parse launchd output, and it also turns
  red for a watcher that hangs.
- Picker: the footer hint lists `/ search`.
- `docs/instruction.md`: troubleshooting entry for a watcher that macOS
  refuses to start after a rebuild (`OS_REASON_CODESIGNING`).

### Changed

- Picker: the filter moves to the top left, on the same line as the entry
  count, with a blank line before the list.
- Picker: the footer is pinned to the last row of the popup.
- Picker: the filter box has no placeholder text.

### Internal

- `picker.go`: comments only on declarations, no inline comments. `View` and
  `reloadIfChanged` are split into `rows`, `footer`, `renderRow`,
  `backgroundOn`, `clearFilter`, `selectedText` and `followText`.
- `datadir_test.go`: test for the heartbeat check.

## 0.3.1 - 2026-10-12

### Changed

- Picker: the selected entry has a `#313244` background across the whole
  row (preview, age and size line). Before, the highlight stopped after
  the first styled piece.
- Picker: the popup paints `#191824` over its whole area, matching
  herdr's frame instead of leaving a differently colored band.
- Picker: a faint `─` rule separates entries, drawn between entries and
  not after the last one.
- Picker: dropped the "clipboard" header title, since herdr already shows
  it in the popup frame. Only the entry count remains.

### Internal

- Renamed row variables in `View()` (`previewLine`, `metaLine`, `fill`,
  `dimSt`, `dotSt`).
- Removed the unused `bold` style and `padRight` helper.

## 0.3.0 - 2026-10-12

### Added

- The watcher registers itself through SMAppService in `Clipwatch.app`.
- `scripts/check.sh`, linters, a pre-commit hook and a CI workflow.

## 0.0.1 - 2026-10-12

### Added

- Clipboard history for macOS as a herdr plugin: a watcher that records
  copies and a popup picker that copies an entry back.
