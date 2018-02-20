package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sahib/brig/cmd/tabwriter"

	"github.com/sahib/brig/client"
	"github.com/urfave/cli"
	yml "gopkg.in/yaml.v2"
)

func handleOffline(ctx *cli.Context, ctl *client.Client) error {
	return ctl.Disconnect()
}

func handleOnline(ctx *cli.Context, ctl *client.Client) error {
	return ctl.Connect()
}

func handleIsOnline(ctx *cli.Context, ctl *client.Client) error {
	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	if self.IsOnline {
		fmt.Println(color.GreenString("online"))
	} else {
		fmt.Println(color.RedString("offline"))
	}

	return nil
}

func handleOnlinePeers(ctx *cli.Context, ctl *client.Client) error {
	infos, err := ctl.OnlinePeers()
	if err != nil {
		return err
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	if len(infos) == 0 {
		fmt.Println("Remote list is empty. Nobody there to ping.")
		return nil
	}

	fmt.Fprintln(tabW, "NAME\tADDR\tROUNDTRIP\tLASTSEEN\t")

	for _, info := range infos {
		suffix := ""
		if info.Err == nil {
			suffix = fmt.Sprintf(
				"%s\t%s",
				info.Roundtrip,
				color.GreenString("✔ "+info.LastSeen.Format(time.Stamp)),
			)
		} else {
			suffix = fmt.Sprintf(
				"∞\t%s",
				color.RedString("✘ "+info.Err.Error()),
			)
		}

		shortAddr := info.Addr
		if len(shortAddr) > 9 {
			shortAddr = shortAddr[:9]
		}

		fmt.Fprintf(tabW, "%s\t%s\t%s\t\n", info.Name, shortAddr, suffix)
	}

	return tabW.Flush()
}

const (
	remoteHelpText = `# No remotes yet. Uncomment the next lines for an example:
# - Name: alice@wonderland.com
#   Fingerprint: QmVA5j2JHPkDTHgZ[...]:SEfXUDeJA1toVnP[...]
`
)

func remoteListToYml(remotes []client.Remote) ([]byte, error) {
	if len(remotes) == 0 {
		// Provide a helpful description, instead of an empty list.
		return []byte(remoteHelpText), nil
	}

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

	fmt.Println(string(data))
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

func handleNetLocate(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()
	timeoutSec := ctx.Int("timeout")

	candidates, err := ctl.NetLocate(who, timeoutSec)
	if err != nil {
		return fmt.Errorf("Failed to locate peers: %v", err)
	}

	if len(candidates) == 0 {
		fmt.Println(color.YellowString("Nothing found."))
		return nil
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	fmt.Fprintln(tabW, "ADDR\tFINGERPRINT\tTYPE\t")
	for _, candidate := range candidates {
		fmt.Printf(
			"%s\t%s\t%s\t\n",
			color.GreenString(candidate.Addr),
			color.YellowString(candidate.Fingerprint),
			strings.Join(candidate.Mask, "|"),
		)
	}

	return tabW.Flush()
}

func handleRemotePing(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()

	msg := fmt.Sprintf("ping to %s: ", color.MagentaString(who))
	roundtrip, err := ctl.RemotePing(who)
	if err != nil {
		msg += color.RedString("✘")
		msg += fmt.Sprintf(" (%v)", err)
	} else {
		msg += color.GreenString("✔")
		msg += fmt.Sprintf(" (%3.5fms)", roundtrip)
	}

	fmt.Println(msg)
	return nil
}

func handlePin(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	return ctl.Pin(path)
}

func handleUnpin(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	return ctl.Unpin(path)
}

func handleWhoami(ctx *cli.Context, ctl *client.Client) error {
	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	printFingerprint := ctx.Bool("fingerprint")
	printName := ctx.Bool("name")

	userName := color.YellowString(self.CurrentUser)
	ownerName := color.GreenString(self.Owner)

	if !printFingerprint && !printName {
		if self.CurrentUser != self.Owner {
			fmt.Printf("# viewing %s's data currently", userName)
		}

		fmt.Printf("- Name: %s\n", userName)
		fmt.Printf("  Fingerprint: %s\n", self.Fingerprint)

		return nil
	}

	if printName {
		fmt.Printf("%s", ownerName)
	}

	if printFingerprint {
		if printName {
			fmt.Printf(" ")
		}
		fmt.Printf("%s", self.Fingerprint)
	}

	fmt.Printf("\n")
	return nil
}
