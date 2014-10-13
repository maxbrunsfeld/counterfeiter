#!/usr/bin/env bash

set -e

counterfeiter='/tmp/counterfeiter_test'

ln -fs $(pwd)/fixtures /tmp/symlinked_fixtures

go build -o $counterfeiter

$counterfeiter fixtures Something
$counterfeiter fixtures HasVarArgs
$counterfeiter fixtures HasImports
$counterfeiter fixtures HasOtherTypes
$counterfeiter fixtures ReusesArgTypes
$counterfeiter fixtures EmbedsInterfaces
$counterfeiter fixtures/aliased_package InAliasedPackage
$counterfeiter /tmp/symlinked_fixtures Something

go build ./fixtures/...

go test -race -v .

rm /tmp/symlinked_fixtures