#!/bin/bash

set -euo pipefail

# Run the tests (with colorful gotest if available):
GOTEST="go test"
if command -v gotest > /dev/null; then
    GOTEST="gotest"
fi

if [ "$#" == 0 ]; then
    $GOTEST -v -parallel 20 ./... 2>&1 | tee log
else
    $GOTEST -v -parallel 20 "$@"  2>&1 | tee log
fi
