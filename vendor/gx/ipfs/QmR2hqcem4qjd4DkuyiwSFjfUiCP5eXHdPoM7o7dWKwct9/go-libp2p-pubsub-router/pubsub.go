package namesys

import (
	"bytes"
	"context"
	"sync"
	"time"

	dshelp "gx/ipfs/QmNP2u7bofwUQptHQGPfabGWtTCbxhNLSZKqbf1uzsup9V/go-ipfs-ds-help"
	u "gx/ipfs/QmPdKqUcHGFdeSpvjVoaTRPPstGif9GBZb5Q56RVw9o69A/go-ipfs-util"
	routing "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing"
	ropts "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing/options"
	floodsub "gx/ipfs/QmSPD4WJu73TE4eJgzbZQTpmfyT5hsh3SEsZnpBAXpaBDA/go-libp2p-floodsub"
	record "gx/ipfs/QmVsp2KdPYE6M8ryzCk5KHLo3zprcY5hBDaYx6uPCFUdxA/go-libp2p-record"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	cid "gx/ipfs/QmapdYm1b22Frv3k17fqrBYTFRxwiaVJkB299Mfn33edeB/go-cid"
	p2phost "gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	logging "gx/ipfs/QmcVVHfdyv15GVPk7NrxdWjh2hLVccXnoD8j2tyQShiXJb/go-log"
	ds "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore"
	dssync "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore/sync"
)

var log = logging.Logger("pubsub-valuestore")

type PubsubValueStore struct {
	ctx  context.Context
	ds   ds.Datastore
	host p2phost.Host
	cr   routing.ContentRouting
	ps   *floodsub.PubSub

	// Map of keys to subscriptions.
	//
	// If a key is present but the subscription is nil, we've bootstrapped
	// but haven't subscribed.
	mx   sync.Mutex
	subs map[string]*floodsub.Subscription

	Validator record.Validator
}

// NewPubsubPublisher constructs a new Publisher that publishes IPNS records through pubsub.
// The constructor interface is complicated by the need to bootstrap the pubsub topic.
// This could be greatly simplified if the pubsub implementation handled bootstrap itself
func NewPubsubValueStore(ctx context.Context, host p2phost.Host, cr routing.ContentRouting, ps *floodsub.PubSub, validator record.Validator) *PubsubValueStore {
	return &PubsubValueStore{
		ctx:       ctx,
		ds:        dssync.MutexWrap(ds.NewMapDatastore()),
		host:      host, // needed for pubsub bootstrap
		cr:        cr,   // needed for pubsub bootstrap
		ps:        ps,
		Validator: validator,
		subs:      make(map[string]*floodsub.Subscription),
	}
}

// Publish publishes an IPNS record through pubsub with default TTL
func (p *PubsubValueStore) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) error {
	p.mx.Lock()
	_, bootstraped := p.subs[key]

	if !bootstraped {
		p.subs[key] = nil
		p.mx.Unlock()

		bootstrapPubsub(p.ctx, p.cr, p.host, key)
	} else {
		p.mx.Unlock()
	}

	log.Debugf("PubsubPublish: publish value for key", key)
	return p.ps.Publish(key, value)
}

func (p *PubsubValueStore) isBetter(key string, val []byte) bool {
	if p.Validator.Validate(key, val) != nil {
		return false
	}

	old, err := p.getLocal(key)
	if err != nil {
		// If the old one is invalid, the new one is *always* better.
		return true
	}

	// Same record. Possible DoS vector, should consider failing?
	if bytes.Equal(old, val) {
		return true
	}

	i, err := p.Validator.Select(key, [][]byte{val, old})
	return err == nil && i == 0
}

func (p *PubsubValueStore) Subscribe(key string) error {
	p.mx.Lock()
	// see if we already have a pubsub subscription; if not, subscribe
	sub := p.subs[key]
	p.mx.Unlock()

	if sub != nil {
		return nil
	}

	// Ignore the error. We have to check again anyways to make sure the
	// record hasn't expired.
	//
	// Also, make sure to do this *before* subscribing.
	p.ps.RegisterTopicValidator(key, func(ctx context.Context, msg *floodsub.Message) bool {
		return p.isBetter(key, msg.GetData())
	})

	sub, err := p.ps.Subscribe(key)
	if err != nil {
		p.mx.Unlock()
		return err
	}

	p.mx.Lock()
	existingSub, bootstraped := p.subs[key]
	if existingSub != nil {
		p.mx.Unlock()
		sub.Cancel()
		return nil
	}

	p.subs[key] = sub
	ctx, cancel := context.WithCancel(p.ctx)
	go p.handleSubscription(sub, key, cancel)
	p.mx.Unlock()

	log.Debugf("PubsubResolve: subscribed to %s", key)

	if !bootstraped {
		// TODO: Deal with publish then resolve case? Cancel behaviour changes.
		go bootstrapPubsub(ctx, p.cr, p.host, key)
	}
	return nil
}

func (p *PubsubValueStore) getLocal(key string) ([]byte, error) {
	dsval, err := p.ds.Get(dshelp.NewKeyFromBinary([]byte(key)))
	if err != nil {
		// Don't invalidate due to ds errors.
		if err == ds.ErrNotFound {
			err = routing.ErrNotFound
		}
		return nil, err
	}
	val := dsval.([]byte)

	// If the old one is invalid, the new one is *always* better.
	if err := p.Validator.Validate(key, val); err != nil {
		return nil, err
	}
	return val, nil
}

func (p *PubsubValueStore) GetValue(ctx context.Context, key string, opts ...ropts.Option) ([]byte, error) {
	if err := p.Subscribe(key); err != nil {
		return nil, err
	}

	return p.getLocal(key)
}

// GetSubscriptions retrieves a list of active topic subscriptions
func (p *PubsubValueStore) GetSubscriptions() []string {
	p.mx.Lock()
	defer p.mx.Unlock()

	var res []string
	for sub := range p.subs {
		res = append(res, sub)
	}

	return res
}

// Cancel cancels a topic subscription; returns true if an active
// subscription was canceled
func (p *PubsubValueStore) Cancel(name string) bool {
	p.mx.Lock()
	defer p.mx.Unlock()

	sub, ok := p.subs[name]
	if ok {
		sub.Cancel()
		delete(p.subs, name)
	}

	return ok
}

func (p *PubsubValueStore) handleSubscription(sub *floodsub.Subscription, key string, cancel func()) {
	defer sub.Cancel()
	defer cancel()

	for {
		msg, err := sub.Next(p.ctx)
		if err != nil {
			if err != context.Canceled {
				log.Warningf("PubsubResolve: subscription error in %s: %s", key, err.Error())
			}
			return
		}
		if p.isBetter(key, msg.GetData()) {
			err := p.ds.Put(dshelp.NewKeyFromBinary([]byte(key)), msg.GetData())
			if err != nil {
				log.Warningf("PubsubResolve: error writing update for %s: %s", key, err)
			}
		}
	}
}

// rendezvous with peers in the name topic through provider records
// Note: rendezvous/boostrap should really be handled by the pubsub implementation itself!
func bootstrapPubsub(ctx context.Context, cr routing.ContentRouting, host p2phost.Host, name string) {
	topic := "floodsub:" + name
	hash := u.Hash([]byte(topic))
	rz := cid.NewCidV1(cid.Raw, hash)

	err := cr.Provide(ctx, rz, true)
	if err != nil {
		log.Warningf("bootstrapPubsub: error providing rendezvous for %s: %s", topic, err.Error())
	}

	go func() {
		for {
			select {
			case <-time.After(8 * time.Hour):
				err := cr.Provide(ctx, rz, true)
				if err != nil {
					log.Warningf("bootstrapPubsub: error providing rendezvous for %s: %s", topic, err.Error())
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	rzctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	wg := &sync.WaitGroup{}
	for pi := range cr.FindProvidersAsync(rzctx, rz, 10) {
		if pi.ID == host.ID() {
			continue
		}
		wg.Add(1)
		go func(pi pstore.PeerInfo) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			err := host.Connect(ctx, pi)
			if err != nil {
				log.Debugf("Error connecting to pubsub peer %s: %s", pi.ID, err.Error())
				return
			}

			// delay to let pubsub perform its handshake
			time.Sleep(time.Millisecond * 250)

			log.Debugf("Connected to pubsub peer %s", pi.ID)
		}(pi)
	}

	wg.Wait()
}
