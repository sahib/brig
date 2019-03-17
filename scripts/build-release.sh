#!/bin/sh

set -e

BIN_DIR="/tmp/brig-binaries"
PLATFORMS="linux/amd64 linux/386 linux/arm darwin/amd64 darwin/386 freebsd/arm freebsd/386 freebsd/amd64"

mkdir -p "$BIN_DIR"

build_all() {
    rm -rf "$BIN_DIR"

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
}

compress_binaries() {
    for bin_path in $(find ${BIN_DIR} -type f -executable); do
        echo "-- Compressing ${bin_path}"

        # upx is pretty happy to return an bad exit code
        # also when the file was packed already.
        set +e
        upx -q $bin_path > /dev/null
        set -e
    done

}

build_checksums() {
    echo "-- Building checksums"
    for bin_path in $(find ${BIN_DIR} -type f -executable); do
        local checksum=$(sha256sum ${bin_path} | cut -d ' ' -f 1)
        echo $checksum > "${bin_path}.sha256"
    done
}

build_archives() {
    for bin_path in $(find ${BIN_DIR} -type f -executable); do
        echo "-- Taring ${bin_path}"
        tar -czf "${bin_path}.tar.gz" \
            -C ${BIN_DIR} \
            $(basename "${bin_path}") \
            $(basename "${bin_path}.sha256") &> /dev/null

        rm -f "${bin_path}" "${bin_path}.sha256"
    done
}

build_all

# This step seems to break some platforms:
# better be safe than sorry.
# compress_binaries

build_checksums
build_archives
echo "-- Binaries are in $BIN_DIR"
