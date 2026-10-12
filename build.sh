#!/bin/sh
#
# build.sh - build Clipwatch.app, the signed app bundle that holds clipwatch.
#
# Usage:
#   ./build.sh
#
# Description:
#   Go compiles only Go, C and Objective-C, so the Swift code is built first.
#   The script:
#     1. compiles swift/*.swift into build/libpasteboard.a, which cgo links;
#     2. builds the Go binary into the bundle;
#     3. copies Info.plist, the watcher launchd plist and the icon from app/;
#     4. writes the version into the bundle's Info.plist (see Versioning);
#     5. signs the bundle ad hoc, which macOS requires.
#   herdr runs this script through the [[build]] entry in herdr-plugin.toml.
#
# Versioning:
#   herdr-plugin.toml is the only place that holds the version. The script
#   copies it to CFBundleShortVersionString, and sets CFBundleVersion to the
#   number of commits in git history (1 when git history is not available).
#   Use scripts/release.sh to change the version.
#
# Requirements:
#   Go and the Xcode Command Line Tools (swiftc, codesign).
#
# Output:
#   build/libpasteboard.a, Clipwatch.app

set -e

cd "$(dirname "$0")"

APP=Clipwatch.app/Contents
VERSION=$(sed -n 's/^version = "\(.*\)"$/\1/p' herdr-plugin.toml | head -n 1)
BUILD=$(git rev-list --count HEAD 2>/dev/null || echo 1)

if [ -z "$VERSION" ]; then
  echo "build.sh: no version found in herdr-plugin.toml" >&2
  exit 1
fi

mkdir -p build "$APP/MacOS" "$APP/Resources" "$APP/Library/LaunchAgents"

swiftc -O -parse-as-library -emit-library -static -module-name pasteboard \
  -o build/libpasteboard.a swift/pasteboard.swift swift/service.swift
go build -o "$APP/MacOS/clipwatch" .

cp app/Info.plist "$APP/Info.plist"
cp app/gusilva.clipwatch.watcher.plist "$APP/Library/LaunchAgents/"
cp app/Clipwatch.icns "$APP/Resources/Clipwatch.icns"

/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $VERSION" "$APP/Info.plist"
/usr/libexec/PlistBuddy -c "Set :CFBundleVersion $BUILD" "$APP/Info.plist"

codesign --force --sign - Clipwatch.app
