#!/bin/sh

set -e

BIN_DIR="/tmp/brig-binaries"
PLATFORMS="linux/amd64 linux/386 linux/arm darwin/amd64 darwin/386 freebsd/arm freebsd/386 freebsd/amd64"

mkdir -p "$BIN_DIR"

build_all() {
    for platform in ${PLATFORMS}; do
        local os=${platform%/*}
        local arch=${platform#*/}
		local log_out="/tmp/build_${log}_${arch}.log"

        echo "-- Building brig_${os}_${arch}"
        BRIG_BINARY_PATH="${BIN_DIR}/brig_${os}_${arch}" GOOS=$os GOARCH=$arch make build-small &> ${log_out}
        if [ $? != 0 ]; then
            echo "-- FAILED:"
            cat ${log_out}
        fi

		rm -f $log_out
    done

    echo "-- Binaries are in $BIN_DIR"
}

build_all
