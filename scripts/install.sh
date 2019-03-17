#!/bin/sh
# This script will download the latest release of brig in your current
# working directory. It also checks if the checksum is the one that
# was advertised. It's meant as quick and dirty installation utility.
# If you really want security, please also check the checksum on the release page:
# https://github.com/sahib/brig/releases

set -e

# NOTE: This script was not tested for many platforms yet.
# It probably needs a bit of adjustment in some cases.

UNAME_OS=$(uname -s)
UNAME_ARCH=$(uname -m)

# We have to map the uname names
# to the names that `go env` uses.
# (without using go env)
GO_OS_NAME=
GO_ARCH_NAME=

# See also:
# https://en.wikipedia.org/wiki/Uname
case "${UNAME_OS}" in
    Linux*)     GO_OS_NAME=linux;;
    Darwin*)    GO_OS_NAME=darwin;;
    FreeBSD*)   GO_OS_NAME=freebsd;;
    *)          echo "The operating system »${UNAME_OS}« is not supported. Sorry."; exit 1
esac

case "${UNAME_ARCH}" in
    x86_64)     GO_ARCH_NAME=amd64;;
    i386*)      GO_ARCH_NAME=386;;
    i686*)      GO_ARCH_NAME=386;;
    armv6*)     GO_ARCH_NAME=arm;;
    armv7*)     GO_ARCH_NAME=arm;;
    armv8*)     GO_ARCH_NAME=arm64;;
    *)          echo "The architecture »${UNAME_ARCH}« is not supported. Sorry."; exit 1
esac

echo "-- Will download binary for ${GO_OS_NAME} and ${GO_ARCH_NAME}."

# Ask GitHub what the latest release is:
RELEASE_METADATA_PATH=$(mktemp)
curl -s https://api.github.com/repos/sahib/brig/releases/latest > ${RELEASE_METADATA_PATH}

# Parse the release URL. This is a bit hacky and would be done nicer via jq,
# but it's not very unlikely to be installed so better not use it.
RELEASE_URL=$( \
    # This is unique to an asset download:
    grep browser_download_url ${RELEASE_METADATA_PATH} | \
    # Extract the url itself:
    grep -o 'https://.*.tar.gz' | \
    # Pick the right OS/ARCH:
    grep "brig_${GO_OS_NAME}_${GO_ARCH_NAME}.tar.gz" | \
    # Make sure to always select the newest version:
    head -1 \
)

rm -f ${RELEASE_METADATA_PATH}
echo "-- Will attempt download from ${RELEASE_URL}"

# Actually download the release now:
DOWNLOAD_ARCHIVE_PATH="$(mktemp --suffix '.brig-release.tar.gz')"
EXTRACTION_PATH="$(mktemp -d --suffix '.brig-extract')"

curl --progress-bar -L ${RELEASE_URL} -o "${DOWNLOAD_ARCHIVE_PATH}"

echo "-- Extracing to ${EXTRACTION_PATH}"
tar xf ${DOWNLOAD_ARCHIVE_PATH} -C ${EXTRACTION_PATH}

BINARY_PATH=$(find ${EXTRACTION_PATH} -type f -executable)
ACTUAL_CHECKSUM=$(sha256sum ${BINARY_PATH} | cut -d ' ' -f 1)
RELEASE_CHECKSUM=$(find ${EXTRACTION_PATH} -type f -iname '*.sha256' -exec cat {} \;)


if [ "${ACTUAL_CHECKSUM}" == "${RELEASE_CHECKSUM}" ]; then
    echo "-- Checksum looks good (${ACTUAL_CHECKSUM})"
else
    echo "-- Checksums are not equal!"
    exit 1
fi

# Copy the actual binary to the current directory if all is fine:
echo "-- Copying binary to ./brig"
cp ${BINARY_PATH} brig

echo "-- Cleaning up unused files"
rm -f ${DOWNLOAD_ARCHIVE_PATH}
rm -rf ${EXTRACTION_PATH}

echo "-- All good. Execute »./brig --help« to read the help or issue the following command to install:"
echo "                                  "
echo "   $ sudo cp ./brig /usr/local/bin"
