#!/bin/bash

set -euo pipefail

sh -c "$(curl -ssL https://taskfile.dev/install.sh)" -- -d

if [ "$EUID" -ne 0 ]; then
    echo '-- Not running as root, so putting binary in ./bin/task'
else
    cp bin/task /usr/bin
    rm -rf bin
fi
