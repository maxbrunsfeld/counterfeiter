#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."
pwd

GOCOMMAND="go"
# GOCOMMAND="go1.22rc1"

# run ${GOCOMMAND} vet to verify everything builds and to check common issues
echo
echo "Running ${GOCOMMAND} vet..."
echo
${GOCOMMAND} vet ./...

# counterfeit all the things
echo
echo "Installing counterfeiter..."
echo
${GOCOMMAND} install .

# counterfeit all the things
echo
echo "Generating fakes used by tests..."
echo
${GOCOMMAND} generate ./...

# validate that the generated fakes match the committed fakes
echo
echo "Validating that generated fakes have not changed..."
echo
git diff --exit-code
if output=$(git status --porcelain) && [ ! -z "$output" ]; then
  echo "the working copy is not clean; make sure that ${GOCOMMAND} generate ./... has been run, and"
  echo "that you have committed or ignored all files before running ./scripts/ci.sh"
  exit 1
fi

# check that the fakes compile
echo
echo "Ensuring generated fakes compile..."
echo
${GOCOMMAND} build -v ./...

# run the tests using the fakes
echo
echo "Running tests..."
echo
${GOCOMMAND} test -race ./...

# run fake generation on transient files
#
# Unfortunately, there seems to be no other good way of doing this. Invalid Go files will
# break external tooling that consumes this repository, so they cannot be stored as Go
# files in version control. Instead, they are stored as .go.txt files and temporarily
# converted into .go files in a "sandbox directory", where `go generate` is run and
# all generated .go files are synced back to the real fixtures directory as .go.txt
# files for dirtiness checking.
echo
echo "Installing temporary fixtures..."
echo

# move fixtures to a temp directory
real_fixtures="$(mktemp -d -t fixtures)"
cp -a fixtures/. "$real_fixtures"
function restore_fixtures {
    EXIT_CODE="$?"
    if [ -d "$real_fixtures" ]; then
        rm -rf fixtures
        mv "$real_fixtures" fixtures
    fi
    exit "$EXIT_CODE"
}
# make sure fixtures are restored if script doesn't succeed
trap restore_fixtures EXIT

# move .txt.go files to .go files in the sandbox fixtures directory
rm -rf fixtures/*
for f in $(find "$real_fixtures" -name "*.go.txt" -exec realpath --relative-to "$real_fixtures" {} \;); do
    mkdir -p "$(dirname "fixtures/$f")"
    cp "$real_fixtures/$f" "fixtures/${f%.txt}"
done

echo
echo "Generating fakes from temporary fixtures..."
echo
GO111MODULE=on go generate ./fixtures/...

echo
echo "Syncing fakes from temporary fixtures..."
echo
# move generate .go files to .txt.go files in the real fixtures directory
for f in $(find "fixtures" -name "*.go" -exec realpath --relative-to "fixtures" {} \;); do
    mv "fixtures/${f}" "$real_fixtures/${f}.txt"
done

# restore the fixtures directory and validate that the generated fake text files match
# the committed fake text files
echo
echo "Validating that temporary fake text files have not changed..."
echo
rm -rf fixtures
mv "$real_fixtures" fixtures
git diff --exit-code
if output=$(git status --porcelain) && [ ! -z "$output" ]; then
  echo "the working copy is not clean; commit or ignore all new .go.txt files"
  echo "before re-running this script"
  exit 1
fi

echo "
 _______  _     _  _______  _______  _______
|       || | _ | ||       ||       ||       |
|  _____|| || || ||    ___||    ___||_     _|
| |_____ |       ||   |___ |   |___   |   |
|_____  ||       ||    ___||    ___|  |   |
 _____| ||   _   ||   |___ |   |___   |   |
|_______||__| |__||_______||_______|  |___|
 _______  __   __  ___   _______  _______
|       ||  | |  ||   | |       ||       |
|  _____||  | |  ||   | |_     _||    ___|
| |_____ |  |_|  ||   |   |   |  |   |___
|_____  ||       ||   |   |   |  |    ___|
 _____| ||       ||   |   |   |  |   |___
|_______||_______||___|   |___|  |_______|
 _______  __   __  _______  _______  _______  _______  _______
|       ||  | |  ||       ||       ||       ||       ||       |
|  _____||  | |  ||       ||       ||    ___||  _____||  _____|
| |_____ |  |_|  ||       ||       ||   |___ | |_____ | |_____
|_____  ||       ||      _||      _||    ___||_____  ||_____  |
 _____| ||       ||     |_ |     |_ |   |___  _____| | _____| |
|_______||_______||_______||_______||_______||_______||_______|
"
