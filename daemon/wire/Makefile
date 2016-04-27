all:
	protoc --proto_path=. --gofast_out . daemon.proto -I ../../store/wire
	sed -i 's#brig_store "."#brig_store "github.com/disorganizer/brig/store/wire"#g' *.go
