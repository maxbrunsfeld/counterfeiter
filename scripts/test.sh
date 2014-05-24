#!/usr/bin/env bash

set -e

go run main.go fixtures Something
ginkgo -race -r
