#!/usr/bin/env bash
set -eu
cd "$(dirname "$0")/.."

go install .
go list ./... | grep -v /vendored | grep -v /generator | xargs go generate
