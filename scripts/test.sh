#!/usr/bin/env bash

set -e

interfaces='Something HasVarArgs'

for interface in $interfaces; do
  go run main.go fixtures $interface
done

go test -v .
