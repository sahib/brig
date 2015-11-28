package daemon

import (
	"fmt"
	"sort"

	log "github.com/Sirupsen/logrus"
	core "github.com/ipfs/go-ipfs/core"
	corenet "github.com/ipfs/go-ipfs/core/corenet"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	"golang.org/x/net/context"
)

func printSwarmAddrs(node *core.IpfsNode) {
	var addrs []string
	for _, addr := range node.PeerHost.Addrs() {
		addrs = append(addrs, addr.String())
	}
	sort.Sort(sort.StringSlice(addrs))

	for _, addr := range addrs {
		fmt.Printf("Swarm listening on %s\n", addr)
	}
}

func startIpfsDaemon() error {
	// Basic ipfsnode setup
	r, err := fsrepo.Open("~/.ipfs")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: true,
	}
	log.Info("NEW NODE")

	nd, err := core.NewNode(ctx, cfg)
	if err != nil {
		return err
	}

	printSwarmAddrs(nd)

	go func() {
		list, err := corenet.Listen(nd, "/app/whyrusleeping")
		if err != nil {
			panic(err)
		}

		fmt.Printf("I am peer: %s\n", nd.Identity.Pretty())

		for {
			con, err := list.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}

			defer con.Close()

			fmt.Fprintln(con, "Hello! This is whyrusleepings awesome ipfs service")
			fmt.Printf("Connection from: %s\n", con.Conn().RemotePeer())
		}

	}()

	return nil
}
