#!/usr/bin/env bash

# Dependencies
go get -v

# Build Static

PLATFORM="$(uname -s | tr 'A-Z' 'a-z')"

export CGO_ENABLED=0
#GOOS=linux
export GOOS=${PLATFORM}
go build -a -ldflags '-extldflags "-static"' -v ./
