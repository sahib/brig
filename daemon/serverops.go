package daemon

import (
	"bytes"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/wire"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/ipfsutil"
	"golang.org/x/net/context"
)

type handlerFunc func(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error)

var handlerMap = map[wire.MessageType]handlerFunc{
	wire.MessageType_ADD:           handleAdd,
	wire.MessageType_CAT:           handleCat,
	wire.MessageType_PING:          handlePing,
	wire.MessageType_QUIT:          handleQuit,
	wire.MessageType_MOUNT:         handleMount,
	wire.MessageType_UNMOUNT:       handleUnmount,
	wire.MessageType_RM:            handleRm,
	wire.MessageType_MV:            handleMv,
	wire.MessageType_HISTORY:       handleHistory,
	wire.MessageType_LOG:           handleLog,
	wire.MessageType_ONLINE_STATUS: handleOnlineStatus,
	wire.MessageType_FETCH:         handleFetch,
	wire.MessageType_LIST:          handleList,
	wire.MessageType_MKDIR:         handleMkdir,
	wire.MessageType_REMOTE_ADD:    handleRemoteAdd,
	wire.MessageType_REMOTE_REMOVE: handleRemoteRemove,
	wire.MessageType_REMOTE_LIST:   handleRemoteList,
	wire.MessageType_REMOTE_LOCATE: handleRemoteLocate,
	wire.MessageType_REMOTE_SELF:   handleRemoteSelf,
}

func handlePing(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	return []byte("PONG"), nil
}

func handleQuit(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	d.signals <- os.Interrupt
	return []byte("BYE"), nil
}

func handleAdd(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	filePath := cmd.GetAddCommand().GetFilePath()
	repoPath := cmd.GetAddCommand().GetRepoPath()

	err := d.Repo.OwnStore.Add(filePath, repoPath)
	if err != nil {
		return nil, err
	}

	return []byte(repoPath), nil
}

func handleCat(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	filePath := cmd.GetCatCommand().GetFilePath()
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	srcPath := cmd.GetCatCommand().GetRepoPath()
	if err := d.Repo.OwnStore.Cat(srcPath, fd); err != nil {
		return nil, err
	}

	return []byte(srcPath), nil
}

func handleMount(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	mountPath := cmd.GetMountCommand().GetMountPoint()

	if _, err := d.Mounts.AddMount(mountPath); err != nil {
		log.Errorf("Unable to mount `%v`: %v", mountPath, err)
		return nil, err
	}

	return []byte(mountPath), nil
}

func handleUnmount(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	mountPath := cmd.GetUnmountCommand().GetMountPoint()

	if err := d.Mounts.Unmount(mountPath); err != nil {
		log.Errorf("Unable to unmount `%v`: %v", mountPath, err)
		return nil, err
	}

	return []byte(mountPath), nil
}

func handleRm(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	rmCmd := cmd.GetRmCommand()
	repoPath := rmCmd.GetRepoPath()
	recursive := rmCmd.GetRecursive()

	if err := d.Repo.OwnStore.Remove(repoPath, recursive); err != nil {
		return nil, err
	}

	return []byte(repoPath), nil
}

func handleMv(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	mvCmd := cmd.GetMvCommand()
	if err := d.Repo.OwnStore.Move(mvCmd.GetSource(), mvCmd.GetDest()); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleHistory(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	repoPath := cmd.GetHistoryCommand().GetRepoPath()

	history, err := d.Repo.OwnStore.History(repoPath)
	if err != nil {
		return nil, err
	}

	protoData, err := history.Marshal()
	if err != nil {
		return nil, err
	}

	return protoData, err
}

func handleLog(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	// TODO: Needs implementation.
	return nil, nil
}

func handleOnlineStatus(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	qry := cmd.GetOnlineStatusCommand().GetQuery()
	switch qry {
	case wire.OnlineQuery_IS_ONLINE:
		if d.IsOnline() {
			return []byte("online"), nil
		} else {
			return []byte("offline"), nil
		}
	case wire.OnlineQuery_GO_ONLINE:
		return nil, d.Connect()
	case wire.OnlineQuery_GO_OFFLINE:
		return nil, d.Disconnect()
	}

	return nil, fmt.Errorf("handleOnlineStatus: Bad query received: %v", qry)
}

func handleFetch(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	// fetchCmd := cmd.GetFetchCommand()
	// who, err := id.Cast(fetchCmd.GetWho())
	// if err != nil {
	// 	return nil, fmt.Errorf("Bad id `%s`: %v", fetchCmd.GetWho(), err)
	// }

	// if !d.MetaHost.IsInOnlineMode() {
	// 	return nil, fmt.Errorf("Metadata Host is not online.")
	// }

	// TODO: Resolve to peer id

	// client, err := d.MetaHost.Dial(who)
	// if err != nil {
	// 	return nil, err
	// }

	// // TODO
	// // importData, err := client.DoFetch()
	// // if err != nil {
	// // 	return nil, err
	// // }

	// if err := d.Repo.OwnStore.Import(bytes.NewReader(importData)); err != nil {
	// 	return nil, err
	// }

	// TODO: what to return on success?
	return []byte("OK"), nil
}

func handleList(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	listCmd := cmd.GetListCommand()
	root, depth := listCmd.GetRoot(), listCmd.GetDepth()
	buf := &bytes.Buffer{}

	if err := d.Repo.OwnStore.ListMarshalled(buf, root, int(depth)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func handleMkdir(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	path := cmd.GetMkdirCommand().GetPath()

	if _, err := d.Repo.OwnStore.Mkdir(path); err != nil {
		return nil, err
	}

	return []byte("OK"), nil
}

func handleRemoteAdd(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	remoteAddCmd := cmd.GetRemoteAddCommand()
	idString, peerHash := remoteAddCmd.GetId(), remoteAddCmd.GetHash()

	id, err := id.Cast(idString)
	if err != nil {
		return nil, err
	}

	remote := repo.NewRemote(id, peerHash)
	if err := d.Repo.Remotes.Insert(remote); err != nil {
		return nil, err
	}

	return []byte("OK"), nil
}

func handleRemoteRemove(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	idString := cmd.GetRemoteRemoveCommand().GetId()

	id, err := id.Cast(idString)
	if err != nil {
		return nil, err
	}

	if err := d.Repo.Remotes.Remove(id); err != nil {
		return nil, err
	}

	return []byte("OK"), nil
}

func handleRemoteList(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	resp := ""
	for idx, rm := range d.Repo.Remotes.List() {
		state := "offline"
		if d.MetaHost.IsOnline(rm) {
			state = "online"
		}

		resp += fmt.Sprintf("#%02d %s: %s\n", idx+1, rm.ID(), rm.Hash(), state)
	}

	return []byte(resp), nil
}

func handleRemoteLocate(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	locateCmd := cmd.GetRemoteLocateCommand()
	idString, peerLimit := locateCmd.GetId(), int(locateCmd.GetPeerLimit())
	timeout := time.Duration(locateCmd.GetTimeoutMs()) * time.Millisecond

	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	id, err := id.Cast(idString)
	if err != nil {
		return nil, err
	}

	peers, err := ipfsutil.Locate(d.Repo.IPFS, id.Hash(), peerLimit, timeout)
	if err != nil {
		return nil, err
	}

	resp := ""
	for _, peer := range peers {
		resp += peer.ID + "\n"
	}

	return []byte(resp), nil
}

func handleRemoteSelf(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	peerHash, err := d.Repo.IPFS.Identity()
	if err != nil {
		return nil, err
	}

	return []byte(peerHash), nil
}
