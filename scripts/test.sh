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
find ./fixtures/ -type d -name '*fakes' | xargs go build

# run the tests using the fakes
go test -race -v ./...

# remove any generated fakes
# this is important because users may have the repo
# checked out for a long time and acquire cruft.
# If they come back and git pull after a long time,
# and some of our internal interfaces have changed,
# they will likely have old generated fakes that reference
# files that no longer exist, breaking their local tests
find ./fixtures/ -type d -name '*fakes' | xargs rm -rf
