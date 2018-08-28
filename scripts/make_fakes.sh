#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."

# build counterfeiter itself
counterfeiter='/tmp/counterfeiter_test'
# shellcheck disable=SC2064
# we want to use the current value
trap "rm $counterfeiter" EXIT
go build -o $counterfeiter

# counterfeit all the interfaces we can find except the version limited ones
egrep --recursive --include '*.go' 'type [^ ]* interface {' . \
      --exclude 'fake_*.go' --exclude '*_test.go' --exclude '*_limited.go' \
  | sed 's#^./\(.*\)/[^/]*.go:type \([^ ]*\) interface {#\1 \2#' \
  | grep -v 'vendor/' \
  | while read -r PACKAGE INTERFACE; do $counterfeiter "$PACKAGE" "$INTERFACE"; done

# counterfeit the limited ones
hasLaterVersion() { test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"; }

currentVersion=$(go version | awk '{print $3}')
if hasLaterVersion ${currentVersion} 'go1.9'; then
     egrep --recursive --include '*_go1.9_limited.go' 'type [^ ]* interface {' . \
      --exclude 'fake_*.go' --exclude '*_test.go' \
  | sed 's#^./\(.*\)/[^/]*.go:type \([^ ]*\) interface {#\1 \2#' \
  | grep -v 'vendor/' \
  | while read -r PACKAGE INTERFACE; do $counterfeiter "$PACKAGE" "$INTERFACE"; done
fi

# fix an import oddity in the UI fake
sed -i.bak '/"golang.org\/x\/crypto\/ssh\/terminal"/d' terminal/terminalfakes/fake_ui.go
