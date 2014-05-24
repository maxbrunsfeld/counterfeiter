#!/usr/bin/env bash

go run main.go fixtures Something && ginkgo -r
