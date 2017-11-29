IMPORT="github.com/disorganizer/brig"

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
			-X $(IMPORT).Major=$(VERSION_MAJOR) \
			-X $(IMPORT).Minor=$(VERSION_MINOR) \
			-X $(IMPORT).Patch=$(VERSION_PATCH) \
			-X $(IMPORT).ReleaseType=$(RELEASETYPE) \
			-X $(IMPORT).BuildTime=$(BUILDTIME) \
			-X $(IMPORT).GitRev=$(GITREV) \
		" \
		cmd/brig/brig.go

test:
	go test -v `glide novendor`

lint:
	gometalinter ./... --deadline 1m | grep -v '.*\.pb\..*'

proto:
	@make -C store/wire
	@make -C daemon/wire
	@make -C transfer/wire
 



