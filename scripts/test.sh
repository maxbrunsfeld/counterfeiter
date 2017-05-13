#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."

# counterfeit all the things
echo
echo "Generating fakes used by tests..."
echo
scripts/make_fakes.sh

# counterfeit through a symlink
symlinked_fixtures=/tmp/symlinked_fixtures
trap "unlink $symlinked_fixtures" EXIT
ln -fs $(pwd)/fixtures $symlinked_fixtures
mkdir -p fixtures/symlinked_fixturesfakes

go run main.go -o fixtures/symlinked_fixturesfakes/fake_something.go $symlinked_fixtures Something

sleep 1

# check that the fakes compile
echo
echo "Ensuring generated fakes compile..."
echo
find ./fixtures/ -type d -name '*fakes' | xargs go build

# run the tests using the fakes
echo
echo "Running tests..."
echo
go test  -v -race ./...

# remove any generated fakes
# this is important because users may have the repo
# checked out for a long time and acquire cruft.
# If they come back and git pull after a long time,
# and some of our internal interfaces have changed,
# they will likely have old generated fakes that reference
# files that no longer exist, breaking their local tests
echo
echo "Removing generated files..."
echo
find ./fixtures/ -type d -name '*fakes/fake*.go' | xargs rm -rf

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

