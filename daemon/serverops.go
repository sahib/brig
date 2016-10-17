package daemon

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/wire"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/store"
	storewire "github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"
)

type handlerFunc func(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error)

var handlerMap = map[wire.MessageType]handlerFunc{
	wire.MessageType_STAGE:         handleStage,
	wire.MessageType_CAT:           handleCat,
	wire.MessageType_PING:          handlePing,
	wire.MessageType_QUIT:          handleQuit,
	wire.MessageType_MOUNT:         handleMount,
	wire.MessageType_UNMOUNT:       handleUnmount,
	wire.MessageType_RM:            handleRm,
	wire.MessageType_MV:            handleMv,
	wire.MessageType_HISTORY:       handleHistory,
	wire.MessageType_ONLINE_STATUS: handleOnlineStatus,
	wire.MessageType_SYNC:          handleSync,
	wire.MessageType_LIST:          handleList,
	wire.MessageType_MKDIR:         handleMkdir,
	wire.MessageType_REMOTE_ADD:    handleRemoteAdd,
	wire.MessageType_REMOTE_REMOVE: handleRemoteRemove,
	wire.MessageType_REMOTE_LIST:   handleRemoteList,
	wire.MessageType_REMOTE_LOCATE: handleRemoteLocate,
	wire.MessageType_REMOTE_SELF:   handleRemoteSelf,
	wire.MessageType_STATUS:        handleStatus,
	wire.MessageType_COMMIT:        handleCommit,
	wire.MessageType_DIFF:          handleDiff,
	wire.MessageType_LOG:           handleLog,
	wire.MessageType_PIN:           handlePin,
	wire.MessageType_EXPORT:        handleExport,
	wire.MessageType_IMPORT:        handleImport,
}

func handlePing(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	return nil, nil
}

func handleQuit(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	d.signals <- os.Interrupt
	return nil, nil
}

func handleStage(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	filePath := cmd.GetAddCommand().FilePath
	repoPath := cmd.GetAddCommand().RepoPath

	err := d.Repo.OwnStore.Stage(filePath, repoPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func handleCat(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	filePath := cmd.GetCatCommand().FilePath
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	srcPath := cmd.GetCatCommand().RepoPath
	if err := d.Repo.OwnStore.Cat(srcPath, fd); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleMount(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	mountPath := cmd.GetMountCommand().MountPoint

	if _, err := d.Mounts.AddMount(mountPath); err != nil {
		log.Errorf("Unable to mount `%v`: %v", mountPath, err)
		return nil, err
	}

	return nil, nil
}

func handleUnmount(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	mountPath := cmd.GetUnmountCommand().MountPoint

	if err := d.Mounts.Unmount(mountPath); err != nil {
		log.Errorf("Unable to unmount `%v`: %v", mountPath, err)
		return nil, err
	}

	return nil, nil
}

func handleRm(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	rmCmd := cmd.GetRmCommand()
	repoPath := rmCmd.RepoPath
	recursive := rmCmd.Recursive

	if err := d.Repo.OwnStore.Remove(repoPath, recursive); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleMv(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	mvCmd := cmd.GetMvCommand()
	if err := d.Repo.OwnStore.Move(mvCmd.Source, mvCmd.Dest, true); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleHistory(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	repoPath := cmd.GetHistoryCommand().RepoPath

	history, err := d.Repo.OwnStore.History(repoPath)
	if err != nil {
		return nil, err
	}

	histProto, err := history.ToProto()
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		HistoryResp: &wire.Response_HistoryResp{
			History: histProto,
		},
	}, nil
}

func handleOnlineStatus(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	qry := cmd.GetOnlineStatusCommand().Query
	switch qry {
	case wire.OnlineQuery_IS_ONLINE:
		return &wire.Response{
			OnlineStatusResp: &wire.Response_OnlineStatusResp{
				IsOnline: d.IsOnline(),
			},
		}, nil
	case wire.OnlineQuery_GO_ONLINE:
		return nil, d.Connect()
	case wire.OnlineQuery_GO_OFFLINE:
		return nil, d.Disconnect()
	}

	return nil, fmt.Errorf("handleOnlineStatus: Bad query received: %v", qry)
}

func handleSync(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	syncCmd := cmd.GetSyncCommand()
	who, err := id.Cast(syncCmd.Who)
	if err != nil {
		return nil, fmt.Errorf("Bad id `%s`: %v", syncCmd.Who, err)
	}

	if !d.MetaHost.IsInOnlineMode() {
		return nil, fmt.Errorf("Metadata Host is not online.")
	}

	client, err := d.MetaHost.DialID(who)
	if err != nil {
		return nil, err
	}

	// This might create a new, empty store if it does not exist yet:
	remoteStore, err := d.Repo.Store(who)
	if err != nil {
		return nil, err
	}

	if err := client.Fetch(remoteStore); err != nil {
		return nil, err
	}

	if err := d.Repo.OwnStore.SyncWith(remoteStore); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleList(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	listCmd := cmd.GetListCommand()
	root, depth := listCmd.Root, listCmd.Depth

	entries, err := d.Repo.OwnStore.ListProtoNodes(root, int(depth))
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		ListResp: &wire.Response_ListResp{
			Entries: entries,
		},
	}, nil
}

func handleMkdir(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	mkdirCmd := cmd.GetMkdirCommand()
	path := mkdirCmd.Path

	var err error
	if mkdirCmd.CreateParents {
		_, err = d.Repo.OwnStore.MkdirAll(path)
	} else {
		_, err = d.Repo.OwnStore.Mkdir(path)
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func handleRemoteAdd(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	remoteAddCmd := cmd.GetRemoteAddCommand()
	idString, peerHash := remoteAddCmd.Id, remoteAddCmd.Hash

	id, err := id.Cast(idString)
	if err != nil {
		return nil, err
	}

	remote := repo.NewRemote(id, peerHash)
	if err := d.Repo.Remotes.Insert(remote); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleRemoteRemove(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	idString := cmd.GetRemoteRemoveCommand().Id

	id, err := id.Cast(idString)
	if err != nil {
		return nil, err
	}

	if err := d.Repo.Remotes.Remove(id); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleRemoteList(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	resp := &wire.Response_RemoteListResp{}

	for _, rm := range d.Repo.Remotes.List() {
		protoRm := &wire.Remote{
			Id:       string(rm.ID()),
			Hash:     rm.Hash(),
			IsOnline: d.MetaHost.IsOnline(rm),
		}

		resp.Remotes = append(resp.Remotes, protoRm)
	}

	return &wire.Response{
		RemoteListResp: resp,
	}, nil
}

func handleRemoteLocate(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	locateCmd := cmd.GetRemoteLocateCommand()
	idString, peerLimit := locateCmd.Id, int(locateCmd.PeerLimit)
	timeout := time.Duration(locateCmd.TimeoutMs) * time.Millisecond

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

	resp := &wire.Response_RemoteLocateResp{}
	for _, peer := range peers {
		resp.Hashes = append(resp.Hashes, peer.ID)
	}

	return &wire.Response{
		RemoteLocateResp: resp,
	}, nil
}

func handleRemoteSelf(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	self := d.Repo.Peer()

	return &wire.Response{
		RemoteSelfResp: &wire.Response_RemoteSelfResp{
			Self: &wire.Remote{
				Id:       string(self.ID()),
				Hash:     self.Hash(),
				IsOnline: d.IsOnline(),
			},
		},
	}, nil
}

func handleStatus(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	status, err := d.Repo.OwnStore.Status()
	if err != nil {
		return nil, err
	}

	pstatus, err := status.ToProto()
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		StatusResp: &wire.Response_StatusResp{
			StageCommit: pstatus,
		},
	}, nil
}

func handleCommit(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	message := cmd.GetCommitCommand().Message

	if err := d.Repo.OwnStore.MakeCommit(message); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleDiff(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	// TODO: Implementation missing.
	return nil, nil
}

func handleLog(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	// TODO: Respect from/to
	nodes, err := d.Repo.OwnStore.Log()
	if err != nil {
		return nil, err
	}

	pnodes := &storewire.Nodes{}
	for _, node := range nodes {
		pnode, err := node.ToProto()
		if err != nil {
			return nil, err
		}

		pnodes.Nodes = append(pnodes.Nodes, pnode)
	}

	return &wire.Response{
		LogResp: &wire.Response_LogResp{
			Nodes: pnodes,
		},
	}, nil
}

func handlePin(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	pinCmd := cmd.GetPinCommand()
	var isPinned bool

	switch balance := pinCmd.Balance; {
	case balance < 0:
		if err := d.Repo.OwnStore.Unpin(pinCmd.Path); err != nil {
			return nil, err
		}

		isPinned = false
	case balance > 0:
		if err := d.Repo.OwnStore.Pin(pinCmd.Path); err != nil {
			return nil, err
		}

		isPinned = true
	case balance == 0:
		var err error
		if isPinned, err = d.Repo.OwnStore.IsPinned(pinCmd.Path); err != nil {
			return nil, err
		}
	}

	return &wire.Response{
		PinResp: &wire.Response_PinResp{
			IsPinned: isPinned,
		},
	}, nil
}

func handleImport(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	pbData := cmd.GetImportCommand().Data

	pbStore := storewire.Store{}
	if err := proto.Unmarshal(pbData, &pbStore); err != nil {
		return nil, err
	}

	if err := d.Repo.OwnStore.Import(&pbStore); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleExport(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	who := cmd.GetExportCommand().Who

	// Figure out the correct store:
	var st *store.Store
	if who == "" {
		st = d.Repo.OwnStore
	} else {
		whoID, err := id.Cast(who)
		if err != nil {
			return nil, err
		}

		st, err = d.Repo.Store(whoID)
		if err != nil {
			return nil, err
		}
	}

	pbStore, err := st.Export()
	if err != nil {
		return nil, err
	}

	data, err := proto.Marshal(pbStore)
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		ExportResp: &wire.Response_ExportResp{
			Data: data,
		},
	}, nil
}
