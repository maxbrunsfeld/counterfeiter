#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."
pwd
find ./ \( -path '*fakes/fake*.go' -o -path './fixtures/samepackage/fake_*.go' -o -path './fixtures/go-hyphenpackage/fake_*.go' \) -print0 | xargs -0 rm -rf
