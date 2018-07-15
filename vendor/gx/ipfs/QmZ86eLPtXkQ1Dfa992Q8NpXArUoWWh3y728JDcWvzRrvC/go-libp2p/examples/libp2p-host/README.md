# The libp2p 'host'

For most applications, the host is the basic building block you'll need to get started. This guide will show how to construct and use a simple host.

The host is an abstraction that manages services on top of a swarm. It provides a clean interface to connect to a service on a given remote peer.

If you want to create a host with a default configuration, you can do the following:

```go
import (
	"context"
	"crypto/rand"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
)


// The context governs the lifetime of the libp2p node
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// To construct a simple host with all the default settings, just use `New`
h, err := libp2p.New(ctx)
if err != nil {
	panic(err)
}

fmt.Printf("Hello World, my hosts ID is %s\n", h.ID())
```

If you want more control over the configuration, you can specify some options to the constructor. For a full list of all the configuration supported by the constructor see: [options.go](https://github.com/libp2p/go-libp2p/blob/master/options.go)

In this snippet we generate our own ID and specified on which address we want to listen:

```go
// Set your own keypair
priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
if err != nil {
	panic(err)
}

h2, err := libp2p.New(ctx,
	// Use your own created keypair
	libp2p.Identity(priv),

	// Set your own listen address
	// The config takes an array of addresses, specify as many as you want.
	libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9000"),
)
if err != nil {
	panic(err)
}

fmt.Printf("Hello World, my second hosts ID is %s\n", h2.ID())
```

And thats it, you have a libp2p host and you're ready to start doing some awesome p2p networking!

In future guides we will go over ways to use hosts, configure them differently (hint: there are a huge number of ways to set these up), and interesting ways to apply this technology to various applications you might want to build.

To see this code all put together, take a look at [host.go](host.go).
