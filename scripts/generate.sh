#!/bin/bash

set -euo pipefail

# Multiline bash array literals are hard...
capnp_paths=()
capnp_paths+=("server/capnp/local_api.capnp")
capnp_paths+=("catfs/nodes/capnp/nodes.capnp")
capnp_paths+=("net/capnp/api.capnp")
capnp_paths+=("catfs/vcs/capnp/patch.capnp")
capnp_paths+=("catfs/capnp/pinner.capnp")
capnp_paths+=("events/capnp/events_api.capnp")
capnp_paths+=("gateway/db/capnp/user.capnp")

go mod download

INCLUDE_PATH="$(go list -f '{{ .Dir }}' zombiezen.com/go/capnproto2)"

for capnp_path in "${capnp_paths[@]}"
do
    echo "-- Generating ${capnp_path}"
    capnp compile \
        -I"${INCLUDE_PATH}/std" \
        -ogo "${capnp_path}"
done


go generate ./...
