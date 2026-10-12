# Changelog

All notable changes to clipwatch. The version lives in `herdr-plugin.toml`.

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
