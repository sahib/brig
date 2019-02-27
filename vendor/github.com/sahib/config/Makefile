lint:
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec gofmt -s -w {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec go fix {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec golint {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec misspell {} \;
	find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' -exec gocyclo -over 20 {} \; | sort -n
