#!/bin/sh

go list -f '{{ join .Imports "\n"}}{{"\n"}}{{ join .TestImports "\n" }}{{"\n"}}{{ join .XTestImports "\n"}}' ./... \
    | grep -v "github.com/maxbrunsfeld/counterfeiter" \
    | grep -E '(github.com)|(golang.org)' \
    | sort \
    | uniq \
    | xargs go get -v
