package cmd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/disorganizer/brig/client"
	"github.com/disorganizer/brig/util/colors"
	"github.com/pksunkara/pygments"
	"github.com/urfave/cli"
	yml "gopkg.in/yaml.v2"
)

func remoteListToYml(remotes []client.Remote) ([]byte, error) {
	// TODO: Provide a nicer representation here.
	return yml.Marshal(remotes)
}

func ymlToRemoteList(data []byte) ([]client.Remote, error) {
	remotes := []client.Remote{}

	if err := yml.Unmarshal(data, &remotes); err != nil {
		return nil, err
	}

	return remotes, nil
}

func handleRemoteAdd(ctx *cli.Context, ctl *client.Client) error {
	remote := client.Remote{
		Name:        ctx.Args().Get(0),
		Fingerprint: ctx.Args().Get(1),
		Folders:     nil,
	}

	if err := ctl.RemoteAdd(remote); err != nil {
		return fmt.Errorf("remote add: %v", err)
	}

	return nil
}

func handleRemoteRemove(ctx *cli.Context, ctl *client.Client) error {
	name := ctx.Args().First()
	if err := ctl.RemoteRm(name); err != nil {
		return fmt.Errorf("remote rm: %v", err)
	}

	return nil
}

func handleRemoteList(ctx *cli.Context, ctl *client.Client) error {
	remotes, err := ctl.RemoteLs()
	if err != nil {
		return fmt.Errorf("remote ls: %v", err)
	}

	if len(remotes) == 0 {
		fmt.Println("None yet. Use `brig remote add <user> <id>` to add some.")
		return nil
	}

	data, err := remoteListToYml(remotes)
	if err != nil {
		return fmt.Errorf("Failed to convert to yml: %v", err)
	}

	// Highlight the yml output (That's more of a joke currently):
	highlighted := pygments.Highlight(string(data), "YAML", "terminal256", "utf-8")
	highlighted = strings.TrimSpace(highlighted)
	fmt.Println(highlighted)
	return nil
}

func handleRemoteEdit(ctx *cli.Context, ctl *client.Client) error {
	remotes, err := ctl.RemoteLs()
	if err != nil {
		return fmt.Errorf("remote ls: %v", err)
	}

	data, err := remoteListToYml(remotes)
	if err != nil {
		return fmt.Errorf("Failed to convert to yml: %v", err)
	}

	// Launch an editor on the received data:
	newData, err := edit(data, "yml")
	if err != nil {
		return fmt.Errorf("Failed to launch editor: %v", err)
	}

	// Save a few network roundtrips if nothing was changed:
	if bytes.Equal(data, newData) {
		fmt.Println("Nothing changed.")
		return nil
	}

	newRemotes, err := ymlToRemoteList(newData)
	if err != nil {
		return err
	}

	if err := ctl.RemoteSave(newRemotes); err != nil {
		return fmt.Errorf("Saving back remotes failed: %v", err)
	}

	return nil
}

func handleRemoteLocate(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()
	candidates, err := ctl.RemoteLocate(who)
	if err != nil {
		return fmt.Errorf("Failed to locate peers: %v", err)
	}

	for _, candidate := range candidates {
		fmt.Println(candidate.Name, candidate.Fingerprint)
	}

	return nil
}

func handleRemotePing(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()

	msg := fmt.Sprintf("ping to %s: ", colors.Colorize(who, colors.Magenta))
	roundtrip, err := ctl.RemotePing(who)
	if err != nil {
		msg += colors.Colorize("✘", colors.Red)
		msg += fmt.Sprintf(" (%v)", err)
	} else {
		msg += colors.Colorize("✔", colors.Green)
		msg += fmt.Sprintf(" (%3.5fms)", roundtrip)
	}

	fmt.Println(msg)
	return nil
}
