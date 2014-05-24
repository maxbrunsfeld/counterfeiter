#!/usr/bin/env bash

set -e

go run main.go fixtures Something
go test -v .
