package daemon

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/wire"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo"
	storewire "github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"
)

type handlerFunc func(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error)

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
	wire.MessageType_ONLINE_STATUS: handleOnlineStatus,
	wire.MessageType_FETCH:         handleFetch,
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

func handleAdd(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	filePath := cmd.GetAddCommand().GetFilePath()
	repoPath := cmd.GetAddCommand().GetRepoPath()

	err := d.Repo.OwnStore.Add(filePath, repoPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func handleCat(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	filePath := cmd.GetCatCommand().GetFilePath()
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	srcPath := cmd.GetCatCommand().GetRepoPath()
	if err := d.Repo.OwnStore.Cat(srcPath, fd); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleMount(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	mountPath := cmd.GetMountCommand().GetMountPoint()

	if _, err := d.Mounts.AddMount(mountPath); err != nil {
		log.Errorf("Unable to mount `%v`: %v", mountPath, err)
		return nil, err
	}

	return nil, nil
}

func handleUnmount(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	mountPath := cmd.GetUnmountCommand().GetMountPoint()

	if err := d.Mounts.Unmount(mountPath); err != nil {
		log.Errorf("Unable to unmount `%v`: %v", mountPath, err)
		return nil, err
	}

	return nil, nil
}

func handleRm(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	rmCmd := cmd.GetRmCommand()
	repoPath := rmCmd.GetRepoPath()
	recursive := rmCmd.GetRecursive()

	if err := d.Repo.OwnStore.Remove(repoPath, recursive); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleMv(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	mvCmd := cmd.GetMvCommand()
	if err := d.Repo.OwnStore.Move(mvCmd.GetSource(), mvCmd.GetDest()); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleHistory(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	repoPath := cmd.GetHistoryCommand().GetRepoPath()

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
	qry := cmd.GetOnlineStatusCommand().GetQuery()
	switch qry {
	case wire.OnlineQuery_IS_ONLINE:
		return &wire.Response{
			OnlineStatusResp: &wire.Response_OnlineStatusResp{
				IsOnline: proto.Bool(d.IsOnline()),
			},
		}, nil
	case wire.OnlineQuery_GO_ONLINE:
		return nil, d.Connect()
	case wire.OnlineQuery_GO_OFFLINE:
		return nil, d.Disconnect()
	}

	return nil, fmt.Errorf("handleOnlineStatus: Bad query received: %v", qry)
}

func handleFetch(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	fetchCmd := cmd.GetFetchCommand()
	who, err := id.Cast(fetchCmd.GetWho())
	if err != nil {
		return nil, fmt.Errorf("Bad id `%s`: %v", fetchCmd.GetWho(), err)
	}

	if !d.MetaHost.IsInOnlineMode() {
		return nil, fmt.Errorf("Metadata Host is not online.")
	}

	client, err := d.MetaHost.DialID(who)
	if err != nil {
		return nil, err
	}

	// TODO: Acutally create the store (in .Store()?)
	remoteStore := d.Repo.Store(who)
	if remoteStore == nil {
		return nil, fmt.Errorf("No store for `%s`", who)
	}

	if err := client.Fetch(remoteStore); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleList(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	listCmd := cmd.GetListCommand()
	root, depth := listCmd.GetRoot(), listCmd.GetDepth()

	dirlist, err := d.Repo.OwnStore.ListProto(root, int(depth))
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		ListResp: &wire.Response_ListResp{
			Dirlist: dirlist,
		},
	}, nil
}

func handleMkdir(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	mkdirCmd := cmd.GetMkdirCommand()
	path := mkdirCmd.GetPath()

	var err error
	if mkdirCmd.GetCreateParents() {
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
	idString, peerHash := remoteAddCmd.GetId(), remoteAddCmd.GetHash()

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
	idString := cmd.GetRemoteRemoveCommand().GetId()

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
			Id:       proto.String(string(rm.ID())),
			Hash:     proto.String(rm.Hash()),
			IsOnline: proto.Bool(d.MetaHost.IsOnline(rm)),
		}

		resp.Remotes = append(resp.Remotes, protoRm)
	}

	return &wire.Response{
		RemoteListResp: resp,
	}, nil
}

func handleRemoteLocate(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
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
				Id:       proto.String(string(self.ID())),
				Hash:     proto.String(self.Hash()),
				IsOnline: proto.Bool(d.IsOnline()),
			},
		},
	}, nil
}

func handleStatus(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	status, err := d.Repo.OwnStore.Status()
	if err != nil {
		return nil, err
	}

	protoCommit, err := status.ToProto()
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		StatusResp: &wire.Response_StatusResp{
			StageCommit: protoCommit,
		},
	}, nil
}

func handleCommit(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	message := cmd.GetCommitCommand().GetMessage()

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
	cmts, err := d.Repo.OwnStore.Log()
	if err != nil {
		return nil, err
	}

	protoCmts, err := cmts.ToProto()
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		LogResp: &wire.Response_LogResp{
			Commits: protoCmts,
		},
	}, nil
}

func handlePin(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	pinCmd := cmd.GetPinCommand()
	var isPinned bool

	switch balance := pinCmd.GetBalance(); {
	case balance < 0:
		if err := d.Repo.OwnStore.Unpin(pinCmd.GetPath()); err != nil {
			return nil, err
		}

		isPinned = false
	case balance > 0:
		if err := d.Repo.OwnStore.Pin(pinCmd.GetPath()); err != nil {
			return nil, err
		}

		isPinned = true
	case balance == 0:
		var err error
		if isPinned, err = d.Repo.OwnStore.IsPinned(pinCmd.GetPath()); err != nil {
			return nil, err
		}
	}

	return &wire.Response{
		PinResp: &wire.Response_PinResp{
			IsPinned: proto.Bool(isPinned),
		},
	}, nil
}

func handleImport(d *Server, ctx context.Context, cmd *wire.Command) (*wire.Response, error) {
	pbData := cmd.GetImportCommand().GetData()

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
	pbStore, err := d.Repo.OwnStore.Export()
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
