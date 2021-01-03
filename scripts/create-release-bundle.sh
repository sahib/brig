#!/bin/bash

set -euo pipefail

PLATFORMS="linux/amd64 linux/386 linux/arm darwin/amd64 darwin/386 freebsd/arm freebsd/386 freebsd/amd64"

BIN_DIR="/tmp/brig-binaries"
rm -rf "$BIN_DIR"
mkdir -p "${BIN_DIR}"

build_all() {
    for platform in ${PLATFORMS}; do
        local os=${platform%/*}
        local arch=${platform#*/}
		local log_out="/tmp/build_${os}_${arch}.log"

        echo "-- Building brig_${os}_${arch}"

        if ! \
            BRIG_BINARY_PATH="${BIN_DIR}/brig_${os}_${arch}" \
            GOOS=$os \
            GOARCH=$arch \
            task --force build &> "${log_out}"; then
            echo "-- FAILED:"
            cat "${log_out}"
        fi

		rm -f "$log_out"
    done
}

build_checksums() {
    echo "-- Building checksums"
    for bin_path in $(find ${BIN_DIR} -type f -executable); do
        local checksum=$(sha256sum ${bin_path} | cut -d ' ' -f 1)
        echo "${checksum}" > "${bin_path}.sha256"
    done
}

build_archives() {
    for bin_path in $(find ${BIN_DIR} -type f -executable); do
        echo "-- Tar-ing ${bin_path}"
        tar -czf "${bin_path}.tar.gz" \
            -C ${BIN_DIR} \
            "$(basename "${bin_path}")" \
            "$(basename "${bin_path}.sha256")" \
        &> /dev/null

        rm -f "${bin_path}" "${bin_path}.sha256"
    done
}

build_all
build_checksums
build_archives

echo "-- Binaries are in $BIN_DIR"
