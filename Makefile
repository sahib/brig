all:
	go install brig/brig.go

test:
	go test $(glide novendor)

lint:
	gometalinter ./... --deadline 1m | grep -v '.*\.pb\..*'

proto:
	@make -C store
	@make -C daemon
	@make -C transfer
