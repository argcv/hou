#!/usr/bin/env bash

# Dependencies
#go get ./cmd/...

# Build Static

PLATFORM="$(uname -s | tr 'A-Z' 'a-z')"

export GO111MODULE=on
export CGO_ENABLED=0
#export GOOS=linux
export GOOS=${PLATFORM}
go build -a -ldflags '-extldflags "-static" -s -w' ./cmd/hou
