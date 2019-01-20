# TODO: Convert this to a go file.
#       make might not be available on all platforms.
#       https://github.com/perkeep/perkeep/blob/master/make.go
VERSION_IMPORT="github.com/sahib/brig/version"

# Build metadata:
VERSION_MAJOR=0
VERSION_MINOR=3
VERSION_PATCH=0
RELEASETYPE=
BUILDTIME=`date -u '+%Y-%m-%dT%H:%M:%S%z'`
GITREV=`git rev-parse HEAD`

all: build

dev: generate build

generate:
	go generate ./...

build:
	time go install -ldflags \
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

install:
ifneq ("$(wildcard brig)","")
	@echo "binary found, installing to /usr/local/bin"
else
	@echo "'brig' binary does not exist; please run 'make build' before 'make install'"
	@exit 1
endif

	@sudo cp brig /usr/local/bin

small:
	time go install -ldflags \
		" -s -w \
			-X $(VERSION_IMPORT).Major=$(VERSION_MAJOR) \
			-X $(VERSION_IMPORT).Minor=$(VERSION_MINOR) \
			-X $(VERSION_IMPORT).Patch=$(VERSION_PATCH) \
			-X $(VERSION_IMPORT).ReleaseType=$(RELEASETYPE) \
			-X $(VERSION_IMPORT).BuildTime=$(BUILDTIME) \
			-X $(VERSION_IMPORT).GitRev=$(GITREV) \
		" \
		brig.go
	upx $(GOPATH)/bin/brig

integration-tests:
	@./test-runner.sh

bob:
	@echo "Running bob as sidekick under brig port :6667 and ipfs port :4003"
	docker run -it -p 4003:4002 -p 6667:6666 brig

docs:
	cd docs && make html
