// Package offline implements IpfsRouting with a client which
// is only able to perform offline operations.
package offline

import (
	"bytes"
	"context"
	"errors"
	"time"

	dshelp "gx/ipfs/QmNP2u7bofwUQptHQGPfabGWtTCbxhNLSZKqbf1uzsup9V/go-ipfs-ds-help"
	routing "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing"
	ropts "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing/options"
	record "gx/ipfs/QmVsp2KdPYE6M8ryzCk5KHLo3zprcY5hBDaYx6uPCFUdxA/go-libp2p-record"
	pb "gx/ipfs/QmVsp2KdPYE6M8ryzCk5KHLo3zprcY5hBDaYx6uPCFUdxA/go-libp2p-record/pb"
	proto "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/proto"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	cid "gx/ipfs/QmapdYm1b22Frv3k17fqrBYTFRxwiaVJkB299Mfn33edeB/go-cid"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	ds "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore"
)

// ErrOffline is returned when trying to perform operations that
// require connectivity.
var ErrOffline = errors.New("routing system in offline mode")

// NewOfflineRouter returns an IpfsRouting implementation which only performs
// offline operations. It allows to Put and Get signed dht
// records to and from the local datastore.
func NewOfflineRouter(dstore ds.Datastore, validator record.Validator) routing.IpfsRouting {
	return &offlineRouting{
		datastore: dstore,
		validator: validator,
	}
}

// offlineRouting implements the IpfsRouting interface,
// but only provides the capability to Put and Get signed dht
// records to and from the local datastore.
type offlineRouting struct {
	datastore ds.Datastore
	validator record.Validator
}

func (c *offlineRouting) PutValue(ctx context.Context, key string, val []byte, _ ...ropts.Option) error {
	if err := c.validator.Validate(key, val); err != nil {
		return err
	}
	if old, err := c.GetValue(ctx, key); err == nil {
		// be idempotent to be nice.
		if bytes.Equal(old, val) {
			return nil
		}
		// check to see if the older record is better
		i, err := c.validator.Select(key, [][]byte{val, old})
		if err != nil {
			// this shouldn't happen for validated records.
			return err
		}
		if i != 0 {
			return errors.New("can't replace a newer record with an older one")
		}
	}
	rec := record.MakePutRecord(key, val)
	data, err := proto.Marshal(rec)
	if err != nil {
		return err
	}

	return c.datastore.Put(dshelp.NewKeyFromBinary([]byte(key)), data)
}

func (c *offlineRouting) GetValue(ctx context.Context, key string, _ ...ropts.Option) ([]byte, error) {
	v, err := c.datastore.Get(dshelp.NewKeyFromBinary([]byte(key)))
	if err != nil {
		return nil, err
	}

	byt, ok := v.([]byte)
	if !ok {
		return nil, errors.New("value stored in datastore not []byte")
	}
	rec := new(pb.Record)
	err = proto.Unmarshal(byt, rec)
	if err != nil {
		return nil, err
	}
	val := rec.GetValue()

	err = c.validator.Validate(key, val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c *offlineRouting) FindPeer(ctx context.Context, pid peer.ID) (pstore.PeerInfo, error) {
	return pstore.PeerInfo{}, ErrOffline
}

func (c *offlineRouting) FindProvidersAsync(ctx context.Context, k *cid.Cid, max int) <-chan pstore.PeerInfo {
	out := make(chan pstore.PeerInfo)
	close(out)
	return out
}

func (c *offlineRouting) Provide(_ context.Context, k *cid.Cid, _ bool) error {
	return ErrOffline
}

func (c *offlineRouting) Ping(ctx context.Context, p peer.ID) (time.Duration, error) {
	return 0, ErrOffline
}

func (c *offlineRouting) Bootstrap(context.Context) error {
	return nil
}

// ensure offlineRouting matches the IpfsRouting interface
var _ routing.IpfsRouting = &offlineRouting{}
