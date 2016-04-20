#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."

# counterfeit all the things
scripts/make_fakes.sh

# counterfeit through a symlink
symlinked_fixtures=/tmp/symlinked_fixtures
trap "unlink $symlinked_fixtures" EXIT
ln -fs $(pwd)/fixtures $symlinked_fixtures
mkdir -p fixtures/symlinked_fixturesfakes
go run main.go -o fixtures/symlinked_fixturesfakes/fake_something.go $symlinked_fixtures Something

# check that the fakes compile
find . -type d -name '*fakes' | xargs go build

# run the tests using the fakes
go test -race -v ./...
