#!/bin/bash
# shellcheck disable=SC2086

set -euo pipefail

# Some linters don't take packages but specific files:
go_files="$(
    find . \
        -type f \
        -iname '*.go'  \
        ! -path '*vendor*' \
        ! -path '*capnp*' \
        ! -iname 'build.go' \
        ! -path '*gateway/static/resource.go' \
)"

# Format and fix common issues:
echo '-- Formatting & auto-fixing things...'
go mod tidy
gofmt -s -w ${go_files}
go fix ./...

echo '-- Running golint'
golint ./... || true

echo '-- Running go vet'
go vet ./... || true

echo '-- Running misspell detector'
misspell -w ${go_files}

echo '-- Running gocyclo'
gocyclo -over 20 ${go_files} | sort -n
