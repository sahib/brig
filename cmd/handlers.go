package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/disorganizer/brig"
	"github.com/disorganizer/brig/brigd/client"
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
	log.Infof("Repository is open now.")
	return nil
}

func handleClose(ctx *cli.Context, ctl *client.Client) error {
	return handleDaemonQuit(ctx, ctl)
}

func handleDaemonPing(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleDaemonWait(ctx *cli.Context) error {
	return nil
}

func handleDaemonQuit(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleDaemon(ctx *cli.Context) error {
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

func handleConfigList(cli *cli.Context) error {
	return nil
}

func handleConfigGet(ctx *cli.Context) error {
	return nil
}

func handleConfigSet(ctx *cli.Context) error {
	return nil
}

func handleInit(ctx *cli.Context) error {
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
