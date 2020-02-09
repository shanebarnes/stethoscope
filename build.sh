#!/bin/bash

set -ex

go vet -v ./...

go build -v cmd/stethoscope/main.go

# Disable test caching; show coverage; enable race detector
go test -count=1 -cover -race -v ./...
