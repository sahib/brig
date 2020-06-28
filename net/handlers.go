package net

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/sahib/brig/net/capnp"
	"github.com/sahib/brig/repo"
	log "github.com/sirupsen/logrus"
)

type requestHandler struct {
	bk             backend.Backend
	rp             *repo.Repository
	ctx            context.Context
	rapi           remotesapi.RemotesAPI
	currRemoteName string
}

func completeExportAllowed(folders []repo.Folder) bool {
	if len(folders) == 0 {
		return true
	}

	for _, folder := range folders {
		if folder.Folder == "/" {
			return true
		}
	}

	return false
}

func (hdl *requestHandler) FetchStore(call capnp.Sync_fetchStore) error {
	// We should only export our complete metadata, when the root directory
	// was enabled or no folders were configured.
	currRemote, err := hdl.rp.Remotes.Remote(hdl.currRemoteName)
	if err != nil {
		return err
	}

	if !completeExportAllowed(currRemote.Folders) {
		log.Warningf("Attempt to read complete store from `%v`", hdl.currRemoteName)
		return errors.New("refusing export")
	}

	fs, err := hdl.rp.FS(hdl.rp.Owner, hdl.bk)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := fs.Export(buf); err != nil {
		return err
	}

	return call.Results.SetData(buf.Bytes())
}

func (hdl *requestHandler) FetchPatch(call capnp.Sync_fetchPatch) error {
	currRemote, err := hdl.rp.Remotes.Remote(hdl.currRemoteName)
	if err != nil {
		return err
	}

	fs, err := hdl.rp.FS(hdl.rp.Owner, hdl.bk)
	if err != nil {
		return err
	}

	// Apply the respective folder filter for this remote.
	prefixes := []string{}
	for _, folder := range currRemote.Folders {
		prefixes = append(prefixes, folder.Folder)
	}

	fromIndex := call.Params.FromIndex()
	fromRev := fmt.Sprintf("commit[%d]", fromIndex)

	log.Debugf("Bundling up all changes starting from: %s", fromRev)
	patchData, err := fs.MakePatchToNext(fromRev, prefixes, currRemote.Name)
	if err != nil {
		return err
	}

	call.Results.SetData(patchData)
	return nil
}

func (hdl *requestHandler) IsCompleteFetchAllowed(call capnp.Sync_isCompleteFetchAllowed) error {
	currRemote, err := hdl.rp.Remotes.Remote(hdl.currRemoteName)
	if err != nil {
		return err
	}

	isAllowed := completeExportAllowed(currRemote.Folders)
	call.Results.SetIsAllowed(isAllowed)
	return nil
}

func (hdl *requestHandler) Ping(call capnp.Meta_ping) error {
	return call.Results.SetReply("ALIVE")
}

func (hdl *requestHandler) Version(call capnp.API_version) error {
	call.Results.SetVersion(1)
	return nil
}

func (hdl *requestHandler) IsPushAllowed(call capnp.Sync_isPushAllowed) error {
	currRemote, err := hdl.rp.Remotes.Remote(hdl.currRemoteName)
	if err != nil {
		return err
	}

	call.Results.SetIsAllowed(currRemote.AcceptPush)
	return nil
}

func (hdl *requestHandler) Push(call capnp.Sync_push) error {
	// NOTE: You might be confused by the name "Push".
	// This is the RECEIVING side of the push.
	currRemote, err := hdl.rp.Remotes.Remote(hdl.currRemoteName)
	if err != nil {
		return err
	}

	if !currRemote.AcceptPush {
		return fmt.Errorf("pushing is not allowed for you")
	}

	log.Infof("Syncing with »%s« because he asked us to via a push.", currRemote.Name)
	return hdl.rapi.Sync(currRemote.Name)
}
