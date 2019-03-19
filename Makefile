# TODO: Convert this to a go file.
#       make might not be available on all platforms.
#       https://github.com/perkeep/perkeep/blob/master/make.go

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION_SPLIT=$(shell sed 's/v\(.*\)\.\(.*\)\.\(.*\)/\1 \2 \3/g' ${ROOT_DIR}/.version)

# Build metadata:
VERSION_MAJOR=$(word 1,${VERSION_SPLIT})
VERSION_MINOR=$(word 2,${VERSION_SPLIT})
VERSION_PATCH=$(word 3,${VERSION_SPLIT})

# alpha, beta and so on:
RELEASETYPE=$(word 4,${VERSION_SPLIT})
BUILDTIME=`date -u '+%Y-%m-%dT%H:%M:%S%z'`
GITREV=`git rev-parse HEAD`

# Where to put the resulting binary:
BRIG_BINARY_PATH ?= ${GOBIN}/brig

# What package contains the version number?
VERSION_IMPORT="github.com/sahib/brig/version"

all: build

dev: generate build

generate:
	go generate ./...

build:
	go build \
		-o "${BRIG_BINARY_PATH}" \
		-ldflags \
		" \
			-X $(VERSION_IMPORT).Major=$(VERSION_MAJOR) \
			-X $(VERSION_IMPORT).Minor=$(VERSION_MINOR) \
			-X $(VERSION_IMPORT).Patch=$(VERSION_PATCH) \
			-X $(VERSION_IMPORT).ReleaseType=$(RELEASETYPE) \
			-X $(VERSION_IMPORT).BuildTime=$(BUILDTIME) \
			-X $(VERSION_IMPORT).GitRev=$(GITREV) \
		" \
		brig.go

test:
	# New go test ignores vendor/
	go test -v ./...

lint:
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec gofmt -s -w {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec go fix {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec golint {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec misspell {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec gocyclo -over 20 {} \; | sort -n
	gosec -exclude=G104 -quiet -fmt json ./... | jq '.Issues[] | select(.file | contains("capnp.go") | not)'

capnp:
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo server/capnp/local_api.capnp
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo catfs/nodes/capnp/nodes.capnp
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo net/capnp/api.capnp
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo catfs/vcs/capnp/patch.capnp
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo catfs/capnp/pinner.capnp
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo events/capnp/events_api.capnp
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo gateway/db/capnp/user.capnp

build-small:
	go build \
		-o "${BRIG_BINARY_PATH}" \
		-ldflags \
		" -s -w \
			-X $(VERSION_IMPORT).Major=$(VERSION_MAJOR) \
			-X $(VERSION_IMPORT).Minor=$(VERSION_MINOR) \
			-X $(VERSION_IMPORT).Patch=$(VERSION_PATCH) \
			-X $(VERSION_IMPORT).ReleaseType=$(RELEASETYPE) \
			-X $(VERSION_IMPORT).BuildTime=$(BUILDTIME) \
			-X $(VERSION_IMPORT).GitRev=$(GITREV) \
		" \
		brig.go
	# upx "${BRIG_BINARY_PATH}"
	@echo "Binary is at ${BRIG_BINARY_PATH}"

docs:
	cd docs && make html

cloc:
	@cloc $(shell find -iname '*.elm' -or -iname '*.go' -a ! -path '*vendor*' ! -path '*capnp*' | head -n -1 | sort | uniq)
