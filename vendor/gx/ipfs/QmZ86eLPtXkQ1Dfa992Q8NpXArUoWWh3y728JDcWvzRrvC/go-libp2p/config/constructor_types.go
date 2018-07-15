package config

import (
	"fmt"
	"reflect"

	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	filter "gx/ipfs/QmSW4uNHbvQia8iZDXzbwjiyHQtnyo9aFqfQAMasj3TJ6Y/go-maddr-filter"
	transport "gx/ipfs/QmW5LxJm2Yo3S7uVsfLM7NsJn2QnKbvZD7uYsZVYR7YViE/go-libp2p-transport"
	pnet "gx/ipfs/QmW7Ump7YyBMr712Ta3iEVh3ZYcfVvJaPryfbCnyE826b4/go-libp2p-interface-pnet"
	mux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
	security "gx/ipfs/QmYMGCwtKbLXfh2MUTbLoP66N9Y24DSg8PFJRG21eE2qwv/go-conn-security"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	host "gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	tptu "gx/ipfs/QmbzWHWFpwCm7GDgrGyKnqUzh48PnRp86VD2Bb8wq6ZfDP/go-libp2p-transport-upgrader"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	crypto "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var (
	// interfaces
	hostType      = reflect.TypeOf((*host.Host)(nil)).Elem()
	networkType   = reflect.TypeOf((*inet.Network)(nil)).Elem()
	transportType = reflect.TypeOf((*transport.Transport)(nil)).Elem()
	muxType       = reflect.TypeOf((*mux.Transport)(nil)).Elem()
	securityType  = reflect.TypeOf((*security.Transport)(nil)).Elem()
	protectorType = reflect.TypeOf((*pnet.Protector)(nil)).Elem()
	privKeyType   = reflect.TypeOf((*crypto.PrivKey)(nil)).Elem()
	pubKeyType    = reflect.TypeOf((*crypto.PubKey)(nil)).Elem()
	pstoreType    = reflect.TypeOf((*pstore.Peerstore)(nil)).Elem()

	// concrete types
	peerIDType   = reflect.TypeOf((peer.ID)(""))
	filtersType  = reflect.TypeOf((*filter.Filters)(nil))
	upgraderType = reflect.TypeOf((*tptu.Upgrader)(nil))
)

var argTypes = map[reflect.Type]constructor{
	upgraderType:  func(h host.Host, u *tptu.Upgrader) interface{} { return u },
	hostType:      func(h host.Host, u *tptu.Upgrader) interface{} { return h },
	networkType:   func(h host.Host, u *tptu.Upgrader) interface{} { return h.Network() },
	muxType:       func(h host.Host, u *tptu.Upgrader) interface{} { return u.Muxer },
	securityType:  func(h host.Host, u *tptu.Upgrader) interface{} { return u.Secure },
	protectorType: func(h host.Host, u *tptu.Upgrader) interface{} { return u.Protector },
	filtersType:   func(h host.Host, u *tptu.Upgrader) interface{} { return u.Filters },
	peerIDType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.ID() },
	privKeyType:   func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore().PrivKey(h.ID()) },
	pubKeyType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore().PubKey(h.ID()) },
	pstoreType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore() },
}

func newArgTypeSet(types ...reflect.Type) map[reflect.Type]constructor {
	result := make(map[reflect.Type]constructor, len(types))
	for _, ty := range types {
		c, ok := argTypes[ty]
		if !ok {
			panic(fmt.Sprintf("missing constructor for type %s", ty))
		}
		result[ty] = c
	}
	return result
}
