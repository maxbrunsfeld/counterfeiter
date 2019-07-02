#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."
echo
echo "Validating that generated fakes have not changed..."
echo
git diff --exit-code
if output=$(git status --porcelain) && [ ! -z "$output" ]; then
  echo "the working copy is not clean; make sure that go generate ./... has been run, and"
  echo "that you have committed or ignored all files before running ./scripts/ci.sh"
  exit 1
fi
