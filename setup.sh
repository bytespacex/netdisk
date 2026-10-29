#!/usr/bin/env bash
# Build netdisk.
#
#   ./build.sh          build for the current platform   -> dist/
#   ./build.sh all      cross-compile every target        -> dist/
#
# The version is stamped into the binary from the date (YY.MM.DD) unless the
# VERSION environment variable is set. See main.go (var Version).
#
# netdisk is pure Go (no cgo), so every target cross-compiles from any host
# with the Go toolchain installed.
set -euo pipefail

cd "$(dirname "$0")"

VERSION="${VERSION:-$(date +%y.%m.%d)}"
LDFLAGS="-s -w -X main.Version=${VERSION}"
OUT="dist"
mkdir -p "$OUT"

build() {
  local os="$1" arch="$2" name_arch="${3:-$2}" ext=""
  if [ "$os" = windows ]; then ext=".exe"; fi
  echo "  build netdisk-${os}_${name_arch}${ext}"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" GOARM=7 \
    go build -trimpath -ldflags="$LDFLAGS" \
    -o "$OUT/netdisk-${os}_${name_arch}${ext}" .
}

echo "netdisk ${VERSION}"

if [ "${1:-}" = "all" ]; then
  build linux   amd64
  build linux   arm64
  build linux   arm   armv7
  build linux   386
  build darwin  amd64
  build darwin  arm64
  build windows amd64
  build windows arm64
  build freebsd amd64
  build freebsd arm64
else
  build "$(go env GOOS)" "$(go env GOARCH)"
fi

echo "done -> $OUT/"
ls -lh "$OUT"

# Ask whether to move artifacts to ~/bn/netdisk (single artifact is renamed) and remove dist/
printf 'Move artifacts to ~/bn/netdisk and remove %s/? [y/N] ' "$OUT"
read -r answer || answer=""
case "$answer" in
  [yY] | [yY][eE][sS])
    dest_dir="$HOME/bn"
    mkdir -p "$dest_dir"
    shopt -s nullglob
    artifacts=("$OUT"/*)
    shopt -u nullglob
    moved=0
    if [ "${#artifacts[@]}" -eq 0 ]; then
      echo "no artifacts found in $OUT/"
    elif [ "${#artifacts[@]}" -eq 1 ]; then
      # Single artifact: move to ~/bn/netdisk with rename
      src="${artifacts[0]}"
      name="netdisk"
      if [ "${src##*.}" = "exe" ]; then name="netdisk.exe"; fi
      target="$dest_dir/$name"
      if [ -d "$target" ]; then
        echo "warning: $target is a directory; keeping $OUT/. Remove it first." >&2
      else
        mv -f "$src" "$target"
        echo "moved -> $target"
        moved=1
      fi
    else
      # Multiple artifacts: move into the ~/bn/netdisk/ directory
      target="$dest_dir/netdisk"
      if [ -e "$target" ] && [ ! -d "$target" ]; then
        echo "warning: $target is a file; keeping $OUT/. Remove it first." >&2
      else
        mkdir -p "$target"
        mv -f "${artifacts[@]}" "$target/"
        echo "moved -> $target/"
        moved=1
      fi
    fi
    if [ "$moved" -eq 1 ]; then
      rm -rf "$OUT"
      echo "removed -> $OUT/"
    else
      echo "kept -> $OUT/"
    fi
    ;;
  *)
    echo "kept -> $OUT/"
    ;;
esac
