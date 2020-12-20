#!/bin/bash

set -euo pipefail

# shellcheck disable=SC2046
cloc $(
    find . -type f \
        -iname '*.go' -or \
        -iname '*.elm' -or \
        -iname '*.sh'  -or \
        -iname 'Dockerfile' \
    | grep -v 'resource.go' | grep -v 'capnp.go' \
)
