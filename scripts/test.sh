#!/usr/bin/env bash

go run main.go fixtures SomeInterface && ginkgo -r
