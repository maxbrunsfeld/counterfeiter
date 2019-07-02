#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."
pwd
find ./ -path '*fakes/fake*.go' -print0 | xargs -0 rm -rf
