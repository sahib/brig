all:
	go install cmd/main/brig.go

lint:
	gometalinter ./... --deadline 1m | grep -v '.*\.pb\..*'
