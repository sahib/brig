package record

import (
	pb "gx/ipfs/QmVsp2KdPYE6M8ryzCk5KHLo3zprcY5hBDaYx6uPCFUdxA/go-libp2p-record/pb"
	proto "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/proto"
)

// MakePutRecord creates a dht record for the given key/value pair
func MakePutRecord(key string, value []byte) *pb.Record {
	record := new(pb.Record)
	record.Key = proto.String(string(key))
	record.Value = value
	return record
}
