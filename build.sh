#!/bin/bash

set -ex

go vet -v ./...

# Disable DWARF debugging information generation
go build -ldflags "-w" -o steth -v cmd/steth/main.go

# Disable test caching; show coverage; enable race detector
go test -count=1 -cover -race -v ./...
