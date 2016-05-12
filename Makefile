all:
	go install -ldflags  "-X main.Major=0 -X main.Minor=1 -X main.Patch=0 -X main.Buildtime=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Gitrev=`git rev-parse HEAD`" brig/brig.go

test:
	go test -v `glide novendor`

lint:
	gometalinter ./... --deadline 1m | grep -v '.*\.pb\..*'

proto:
	@make -C store/wire
	@make -C daemon/wire
	@make -C transfer/wire
