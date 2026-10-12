#!/bin/sh
#
# build.sh - build the clipwatch binary, including its Swift pasteboard library.
#
# Usage:
#   ./build.sh
#
# Description:
#   The clipboard code is written in Swift (swift/pasteboard.swift), but the Go
#   toolchain compiles only Go, C and Objective-C, so a plain "go build" fails
#   at link time. This script:
#     1. compiles the Swift source into build/libpasteboard.a, which cgo links
#        (see pasteboard.go);
#     2. builds the Go binary into bin/clipwatch.
#   herdr's [[build]] entry in herdr-plugin.toml runs a single command, so it
#   runs this script.
#
# Requirements:
#   Go and the Xcode Command Line Tools (swiftc).
#
# Output:
#   build/libpasteboard.a, bin/clipwatch

set -e

cd "$(dirname "$0")"
mkdir -p bin build

swiftc -O -parse-as-library -emit-library -static -module-name pasteboard \
  -o build/libpasteboard.a swift/pasteboard.swift
go build -o bin/clipwatch .
