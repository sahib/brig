#!/bin/bash

set -euo pipefail

# Run the tests (with colorful gotest if available):
GOTEST="go test"
if command -v gotest > /dev/null; then
    GOTEST="gotest"
fi

if [ "$#" == 0 ]; then
    $GOTEST -v -parallel 20 ./...
else
    $GOTEST -v -parallel 20 "$@"
fi
