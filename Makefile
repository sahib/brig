VERSION_IMPORT="github.com/sahib/brig/version"

# Build metadata:
VERSION_MAJOR=0
VERSION_MINOR=1
VERSION_PATCH=0
RELEASETYPE=alpha
BUILDTIME=`date -u '+%Y-%m-%dT%H:%M:%S%z'`
GITREV=`git rev-parse HEAD`

all:
	go install -ldflags \
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

proto:
	@make -C store/wire
	@make -C daemon/wire
	@make -C transfer/wire
