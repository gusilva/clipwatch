#!/bin/sh
#
# check.sh - run every automated check for clipwatch.
#
# Usage:
#   scripts/check.sh
#
# Description:
#   Runs, in order, and stops at the first failure:
#     1. sh -n on every shell script;
#     2. golangci-lint (gofmt, vet, slog, doc comments; see .golangci.yml);
#     3. swiftlint (doc comments; see .swiftlint.yml);
#     4. a docs check that fails on paths that no longer exist;
#     5. ./build.sh, then go test.
#   The pre-commit hook and CI run this script.
#
# Requirements:
#   Go, golangci-lint, swiftlint and the Xcode Command Line Tools.

set -e

cd "$(dirname "$0")/.."

for script in build.sh scripts/*.sh; do
  sh -n "$script"
done

golangci-lint run
swiftlint lint --quiet --strict

STALE='bin/clipwatch|\.local/state/herdr|~/Library/LaunchAgents|HERDR_PLUGIN_STATE_DIR'
if grep -nE "$STALE" README.md docs/*.md; then
  echo "check.sh: docs mention paths that no longer exist (above)" >&2
  exit 1
fi

./build.sh
go test ./...
