#!/usr/bin/env bash
set -eu
cd "$(dirname "$0")/.."

go install .
go generate ./...
