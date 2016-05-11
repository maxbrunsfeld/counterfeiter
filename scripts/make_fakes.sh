#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."

# build counterfeiter itself
counterfeiter='/tmp/counterfeiter_test'
trap "rm $counterfeiter" EXIT
go build -o $counterfeiter

# counterfeit all the interfaces we can find
egrep --recursive --include '*.go' 'type [^ ]* interface {' . \
      --exclude-dir 'expected_output' --exclude '*_test.go' \
  | sed 's#^./\(.*\)/[^/]*.go:type \([^ ]*\) interface {#\1 \2#' \
  | while read PACKAGE INTERFACE; do $counterfeiter $PACKAGE $INTERFACE; done

# fix an import oddity in the UI fake
sed -i.bak '/"golang.org\/x\/crypto\/ssh\/terminal"/d' terminal/terminalfakes/fake_ui.go
