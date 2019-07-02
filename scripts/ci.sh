#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."
pwd

# run go vet to verify everything builds and to check common issues
echo
echo "Running go vet..."
echo
GO111MODULE=on go vet ./...

# counterfeit all the things
echo
echo "Installing counterfeiter..."
echo
GO111MODULE=on go install .

# counterfeit all the things
echo
echo "Generating fakes used by tests..."
echo
GO111MODULE=on go generate ./...

# validate that the generated fakes match the committed fakes
echo
echo "Validating that generated fakes have not changed..."
echo
git diff --exit-code
if output=$(git status --porcelain) && [ ! -z "$output" ]; then
  echo "the working copy is not clean; make sure that go generate ./... has been run, and"
  echo "that you have committed or ignored all files before running ./scripts/ci.sh"
  exit 1
fi

# check that the fakes compile
echo
echo "Ensuring generated fakes compile..."
echo
GO111MODULE=on go build -v ./...

# run the tests using the fakes
echo
echo "Running tests..."
echo
GO111MODULE=on go test -race ./...

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
