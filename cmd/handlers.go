package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/disorganizer/brig"
	"github.com/disorganizer/brig/brigd/client"
	"github.com/disorganizer/brig/brigd/server"
	"github.com/disorganizer/brig/util/colors"
)

const brigLogo = `
          _____                   _____                   _____                   _____          
         /\    \                 /\    \                 /\    \                 /\    \         
        /::\    \               /::\    \               /::\    \               /::\    \        
       /::::\    \             /::::\    \              \:::\    \             /::::\    \       
      /::::::\    \           /::::::\    \              \:::\    \           /::::::\    \      
     /:::/\:::\    \         /:::/\:::\    \              \:::\    \         /:::/\:::\    \     
    /:::/__\:::\    \       /:::/__\:::\    \              \:::\    \       /:::/  \:::\    \    
   /::::\   \:::\    \     /::::\   \:::\    \             /::::\    \     /:::/    \:::\    \   
  /::::::\   \:::\    \   /::::::\   \:::\    \   ____    /::::::\    \   /:::/    / \:::\    \  
 /:::/\:::\   \:::\ ___\ /:::/\:::\   \:::\____\ /\   \  /:::/\:::\    \ /:::/    /   \:::\ ___\ 
/:::/__\:::\   \:::|    /:::/  \:::\   \:::|    /::\   \/:::/  \:::\____/:::/____/  ___\:::|    |
\:::\   \:::\  /:::|____\::/   |::::\  /:::|____\:::\  /:::/    \::/    \:::\    \ /\  /:::|____|
 \:::\   \:::\/:::/    / \/____|:::::\/:::/    / \:::\/:::/    / \/____/ \:::\    /::\ \::/    / 
  \:::\   \::::::/    /        |:::::::::/    /   \::::::/    /           \:::\   \:::\ \/____/  
   \:::\   \::::/    /         |::|\::::/    /     \::::/____/             \:::\   \:::\____\    
    \:::\  /:::/    /          |::| \::/____/       \:::\    \              \:::\  /:::/    /    
     \:::\/:::/    /           |::|  ~|              \:::\    \              \:::\/:::/    /     
      \::::::/    /            |::|   |               \:::\    \              \::::::/    /      
       \::::/    /             \::|   |                \:::\____\              \::::/    /       
        \::/____/               \:|   |                 \::/    /               \::/____/        
         ~~                      \|___|                  \/____/                                 
`

func handleVersion(ctx *cli.Context) error {
	fmt.Println(brig.VersionString())
	return nil
}

func handleOpen(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleClose(ctx *cli.Context, ctl *client.Client) error {
	return handleDaemonQuit(ctx, ctl)
}

func handleDaemonPing(ctx *cli.Context, ctl *client.Client) error {
	for i := 0; i < 100; i++ {
		before := time.Now()
		symbol := colors.Colorize("✔", colors.Green)

		if err := ctl.Ping(); err != nil {
			symbol = colors.Colorize("✘", colors.Red)
		}

		delay := time.Since(before)
		fmt.Printf("#%02d %s ➔ %s: %s (%v)\n",
			i+1,
			ctl.LocalAddr().String(),
			ctl.RemoteAddr().String(),
			symbol,
			delay,
		)

		time.Sleep(1 * time.Second)
	}

	return nil
}

func handleDaemonWait(ctx *cli.Context) error {
	return nil
}

func handleDaemonQuit(ctx *cli.Context, ctl *client.Client) error {
	if err := ctl.Quit(); err != nil {
		return ExitCode{
			DaemonNotResponding,
			fmt.Sprintf("brigd not responding: %v", err),
		}
	}

	return nil
}

func handleDaemonLaunch(ctx *cli.Context) error {
	brigPath := os.Getenv("BRIG_PATH")
	if brigPath == "" {
		// TODO: Check parent directories to see if we're in some
		//       brig repository.
		brigPath = "."
	}

	server, err := server.BootServer(brigPath, guessPort())
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to boot brigd: %v", err),
		}
	}

	defer server.Close()

	if err := server.Serve(); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to serve: %v", err),
		}
	}

	return nil
}

func doMount(ctx *cli.Context, ctl *client.Client, mount bool) error {
	return nil
}

func handleMount(ctx *cli.Context, ctl *client.Client) error {
	return doMount(ctx, ctl, !ctx.Bool("unmount"))
}

func handleUnmount(ctx *cli.Context, ctl *client.Client) error {
	return doMount(ctx, ctl, false)
}

func handleInit(ctx *cli.Context, ctl *client.Client) error {
	// Accumulate args:
	owner := ctx.Args().First()
	folder := guessRepoFolder()
	backend := ctx.String("backend")

	if err := ctl.Init(folder, owner, backend); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("init failed: %v", err)}
	}

	fmt.Println(brigLogo)
	return nil
}

func handleConfigList(cli *cli.Context) error {
	return nil
}

func handleConfigGet(ctx *cli.Context) error {
	return nil
}

func handleConfigSet(ctx *cli.Context) error {
	return nil
}

func handleStage(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleRm(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleCat(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleHistory(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleOffline(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleIsOnline(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleOnline(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleList(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleTree(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleMv(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleMkdir(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleStatus(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleCommit(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleLog(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleDiff(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handlePin(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleUnpin(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleDebugExport(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleDebugImport(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleSync(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleReset(ctx *cli.Context, ctl *client.Client) error {
	return nil
}
