#!/bin/sh
#
# release.sh - set the plugin version, commit it and tag it.
#
# Usage:
#   scripts/release.sh X.Y.Z
#
# Description:
#   A git tag cannot change files, so the version is bumped first and the tag
#   is created from that commit. The script:
#     1. checks that the working tree is clean and the tag does not exist;
#     2. sets version in herdr-plugin.toml, the single source of truth that
#        build.sh copies into Clipwatch.app/Contents/Info.plist;
#     3. commits the change and creates the tag vX.Y.Z.
#   Nothing is pushed. Push with: git push && git push origin vX.Y.Z
#
# Requirements:
#   git, and a clean working tree.

set -e

cd "$(dirname "$0")/.."

VERSION=$1

case "$VERSION" in
  [0-9]*.[0-9]*.[0-9]*) ;;
  *)
    echo "usage: scripts/release.sh X.Y.Z" >&2
    exit 2
    ;;
esac

if [ -n "$(git status --porcelain)" ]; then
  echo "release.sh: working tree is not clean" >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/v$VERSION" >/dev/null; then
  echo "release.sh: tag v$VERSION already exists" >&2
  exit 1
fi

sed -i '' "s/^version = \".*\"$/version = \"$VERSION\"/" herdr-plugin.toml

git add herdr-plugin.toml
git commit -m "chore(release): v$VERSION"
git tag -a "v$VERSION" -m "v$VERSION"
