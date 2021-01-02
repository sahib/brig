#!/bin/bash

set -euo pipefail

# Collect some basic info about the repo state:
BUILD_TIME="$(date --iso-8601=seconds)"
GIT_REVISION="$(git rev-parse HEAD)"
CURRENT_BRANCH="$(git branch --show-current)"

VERSION_PACKAGE='github.com/sahib/brig/version'

# TODO: Parse this from the last git tag.
MAJOR=0
MINOR=5
PATCH=3

# Find out where to put the binary.
BINARY_PATH="${BRIG_BINARY_PATH:-${GOBIN:-${GOPATH:-${HOME:-.}/go}/bin}}"
mkdir -p "${BINARY_PATH}"

echo ".brig. is here: ${BINARY_PATH}"

go build \
  -ldflags " \
    -X ${VERSION_PACKAGE}.Major=${MAJOR} \
    -X ${VERSION_PACKAGE}.Minor=${MINOR} \
    -X ${VERSION_PACKAGE}.Patch=${PATCH} \
    -X ${VERSION_PACKAGE}.ReleaseType=${CURRENT_BRANCH} \
    -X ${VERSION_PACKAGE}.BuildTime=${BUILD_TIME} \
    -X ${VERSION_PACKAGE}.GitRev=${GIT_REVISION} \
    -s \
    -w
" -o "${BINARY_PATH}/brig" .
