package cmdline

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/disorganizer/brig/daemon"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/util/colors"
)

func handleRemoteAdd(ctx *cli.Context, client *daemon.Client) error {
	idString, hash := ctx.Args()[0], ctx.Args()[1]

	id, err := id.Cast(idString)
	if err != nil {
		return ExitCode{
			BadArgs,
			fmt.Sprintf("Invalid ID: %v", err),
		}
	}

	if err := client.RemoteAdd(id, hash); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to add remote: %v", err),
		}
	}

	return nil
}

func handleRemoteRemove(ctx *cli.Context, client *daemon.Client) error {
	idString := ctx.Args()[0]
	id, err := id.Cast(idString)
	if err != nil {
		return ExitCode{
			BadArgs,
			fmt.Sprintf("Invalid ID: %v", err),
		}
	}

	if client.RemoteRemove(id) != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to remove remote: %v", err),
		}
	}

	return nil
}

func handleRemoteList(ctx *cli.Context, client *daemon.Client) error {
	data, err := client.RemoteList()
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to list remotes: %v", err),
		}
	}
	for _, entry := range data {
		printRemoteEntry(entry)
	}
	return nil
}

func handleRemoteLocate(ctx *cli.Context, client *daemon.Client) error {
	id, err := id.Cast(ctx.Args()[0])
	if err != nil {
		return ExitCode{
			BadArgs,
			fmt.Sprintf("Invalid ID: %v", err),
		}
	}

	hashes, err := client.RemoteLocate(id, 10, 50000)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to locate ipfs peers: %v", err),
		}
	}

	for _, hash := range hashes {
		fmt.Println(hash)
	}

	return nil
}

func printRemoteEntry(re *daemon.RemoteEntry) {
	state := colors.Colorize("offline", colors.Red)
	if re.IsOnline {
		state = colors.Colorize("online", colors.Green)
	}

	fmt.Printf("%s %-7s %s\n", re.Hash, state, re.Ident)
}

func handleRemoteSelf(ctx *cli.Context, client *daemon.Client) error {
	re, err := client.RemoteSelf()
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to list remote self information: %v", err),
		}
	}

	printRemoteEntry(re)
	return nil
}
