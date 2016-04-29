package cmdline

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon"
	"github.com/disorganizer/brig/id"
	"github.com/qitta/minilock/colors"
	"github.com/tucnak/climax"
)

func handleRemoteAdd(ctx climax.Context, client *daemon.Client) int {
	if len(ctx.Args) < 3 {
		log.Warningf("Need an id and a peer hash.")
		return BadArgs
	}

	idString, hash := ctx.Args[1], ctx.Args[2]

	id, err := id.Cast(idString)
	if err != nil {
		log.Errorf("Invalid ID: %v", err)
		return BadArgs
	}

	if err := client.RemoteAdd(id, hash); err != nil {
		log.Errorf("Unable to add remote: %v", err)
		return UnknownError
	}

	return Success
}

func handleRemoteRemove(ctx climax.Context, client *daemon.Client) int {
	if len(ctx.Args) < 2 {
		log.Warningf("Need an id of the remote to remove.")
		return BadArgs
	}

	idString := ctx.Args[1]
	id, err := id.Cast(idString)
	if err != nil {
		log.Errorf("Invalid ID: %v", err)
		return BadArgs
	}

	if client.RemoteRemove(id) != nil {
		log.Errorf("Unable to remove remote: %v", err)
		return UnknownError
	}

	return Success
}

func handleRemoteList(ctx climax.Context, client *daemon.Client) int {
	data, err := client.RemoteList()
	if err != nil {
		log.Errorf("Unable to list remotes: %v", err)
		return UnknownError
	}
	for _, entry := range data {
		printRemoteEntry(entry)
	}
	return Success
}

func handleRemoteLocate(ctx climax.Context, client *daemon.Client) int {
	if len(ctx.Args) < 2 {
		return BadArgs
	}

	id, err := id.Cast(ctx.Args[1])
	if err != nil {
		log.Errorf("Invalid ID: %v", err)
		return BadArgs
	}

	hashes, err := client.RemoteLocate(id, 10, 50000)
	if err != nil {
		log.Errorf("Unable to locate ipfs peers: %v", err)
		return UnknownError
	}

	for _, hash := range hashes {
		fmt.Println(hash)
	}

	return Success
}

func printRemoteEntry(re *daemon.RemoteEntry) {
	state := colors.Colorize("offline", colors.Red)
	if re.IsOnline {
		state = colors.Colorize("online", colors.Green)
	}

	fmt.Printf("%s %-7s %s\n", re.Hash, state, re.Ident)
}

func handleRemoteSelf(ctx climax.Context, client *daemon.Client) int {
	re, err := client.RemoteSelf()
	if err != nil {
		log.Errorf("Unable to list remote self information: %v", err)
		return UnknownError
	}

	printRemoteEntry(re)
	return Success
}

func handleRemote(ctx climax.Context, client *daemon.Client) int {
	if len(ctx.Args) == 0 {
		return handleRemoteList(ctx, client)
	}

	switch ctx.Args[0] {
	case "add":
		return handleRemoteAdd(ctx, client)
	case "remove":
		return handleRemoteRemove(ctx, client)
	case "list":
		return handleRemoteList(ctx, client)
	case "locate":
		return handleRemoteLocate(ctx, client)
	case "self":
		return handleRemoteSelf(ctx, client)
	}

	log.Warningf("No remote subcommand `%s`", ctx.Args[0])
	return BadArgs
}
