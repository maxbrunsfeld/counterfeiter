#!/usr/bin/env bash
set -eu
cd "$(dirname "$0")/.."

# build counterfeiter itself
counterfeiter='/tmp/counterfeiter_test'
# shellcheck disable=SC2064
# we want to use the current value
trap "rm $counterfeiter" EXIT
go build -o $counterfeiter

# generate fakes
go generate ./...
