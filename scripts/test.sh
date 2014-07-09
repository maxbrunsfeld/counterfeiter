#!/usr/bin/env bash

set -e

counterfeiter='/tmp/counterfeiter_test'

go build -o $counterfeiter

$counterfeiter fixtures Something
$counterfeiter fixtures HasVarArgs
$counterfeiter fixtures HasImports
$counterfeiter fixtures HasOtherTypes
$counterfeiter fixtures ReusesArgTypes
$counterfeiter fixtures EmbedsInterfaces
$counterfeiter fixtures/another_package InAliasedPackage

go test -race -v .
