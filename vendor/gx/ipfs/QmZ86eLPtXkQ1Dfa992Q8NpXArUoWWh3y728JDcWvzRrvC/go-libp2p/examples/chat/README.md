# p2p chat app with libp2p

This program demonstrates a simple p2p chat application. It can work between two peers if
1. Both have private IP address (same network).
2. At least one of them has a public IP address.

Assume if 'A' and 'B' are on different networks host 'A' may or may not have a public IP address but host 'B' has one.

Usage: Run `./chat -sp <SOURCE_PORT>` on host 'B' where <SOURCE_PORT> can be any port number. Now run `./chat -d <MULTIADDR_B>` on node 'A' [`<MULTIADDR_B>` is multiaddress of host 'B' which can be obtained from host 'B' console].

## Build

To build the example, first run `make deps` in the root directory.

```
> make deps
> go build ./examples/chat
```

## Usage

On node 'B'

```
> ./chat -sp 3001
Run ./chat -d /ip4/127.0.0.1/tcp/3001/ipfs/QmdXGaeGiVA745XorV1jr11RHxB9z4fqykm6xCUPX1aTJo

2018/02/27 01:21:32 Got a new stream!
> hi (received messages in green colour)
> hello (sent messages in white colour)
> no
```

On node 'A'. Replace 127.0.0.1 with <PUBLIC_IP> if node 'B' has one.

```
> ./chat -d /ip4/127.0.0.1/tcp/3001/ipfs/QmdXGaeGiVA745XorV1jr11RHxB9z4fqykm6xCUPX1aTJo
Run ./chat -d /ip4/127.0.0.1/tcp/3001/ipfs/QmdXGaeGiVA745XorV1jr11RHxB9z4fqykm6xCUPX1aTJo

This node's multiaddress:
/ip4/0.0.0.0/tcp/0/ipfs/QmWVx9NwsgaVWMRHNCpesq1WQAw2T3JurjGDNeVNWifPS7
> hi
> hello
```

**NOTE: debug mode is enabled by default, debug mode will always generate same node id (on each node) on every execution. Disable debug using `--debug false` flag while running your executable.**

## Authors
1. Abhishek Upperwal