all:
	go install brig/brig.go

test:
	go test -v `glide novendor`

lint:
	gometalinter ./... --deadline 1m | grep -v '.*\.pb\..*'

proto:
	@make -C store/wire
	@make -C daemon/wire
	@make -C transfer/wire
