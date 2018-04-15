VERSION_IMPORT="github.com/sahib/brig/version"

# Build metadata:
VERSION_MAJOR=0
VERSION_MINOR=1
VERSION_PATCH=0
RELEASETYPE=alpha
BUILDTIME=`date -u '+%Y-%m-%dT%H:%M:%S%z'`
GITREV=`git rev-parse HEAD`

all: build

build:
	go build -ldflags \
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
	gometalinter ./... --deadline 1m | grep -v '.*\.pb\..*'

capnp:
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo server/capnp/api.capnp
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo catfs/nodes/capnp/model.capnp
	capnp compile -I/home/sahib/go/src/zombiezen.com/go/capnproto2/std -ogo net/capnp/api.capnp

install:
ifneq ("$(wildcard brig)","")
	@echo "binary found, installing to /usr/local/bin"
else
	@echo "'brig' binary does not exist; please run 'make build' before 'make install'"
	@exit 1
endif

	@sudo cp brig /usr/local/bin
