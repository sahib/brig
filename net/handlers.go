package net

import (
	"bytes"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/net/capnp"
	"github.com/sahib/brig/repo"
)

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

func (hdl *handler) FetchStore(call capnp.Sync_fetchStore) error {
	// We should only export our complete metadata, when the root directory
	// was enabled or no folders were configured.
	if !completeExportAllowed(hdl.currRemote.Folders) {
		log.Warningf("Attempt to read complete store from `%v`", hdl.currRemote.Name)
		return errors.New("refusing export")
	}

	user := hdl.rp.CurrentUser()
	fs, err := hdl.rp.FS(user, hdl.bk)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := fs.Export(buf); err != nil {
		return err
	}

	return call.Results.SetData(buf.Bytes())
}

func (hdl *handler) FetchPatch(call capnp.Sync_fetchPatch) error {
	user := hdl.rp.CurrentUser()
	fs, err := hdl.rp.FS(user, hdl.bk)
	if err != nil {
		return err
	}

	// Apply the respective folder filter for this remote.
	prefixes := []string{}
	for _, folder := range hdl.currRemote.Folders {
		prefixes = append(prefixes, folder.Folder)
	}

	fromIndex := call.Params.FromIndex()
	fromRev := fmt.Sprintf("commit[%d]", fromIndex)

	patchData, err := fs.MakePatch(fromRev, prefixes)
	if err != nil {
		return err
	}

	call.Results.SetData(patchData)
	return nil
}

func (hdl *handler) IsCompleteFetchAllowed(call capnp.Sync_isCompleteFetchAllowed) error {
	isAllowed := completeExportAllowed(hdl.currRemote.Folders)
	call.Results.SetIsAllowed(isAllowed)
	return nil
}

func (hdl *handler) Ping(call capnp.Meta_ping) error {
	return call.Results.SetReply("ALIVE")
}

func (hdl *handler) Version(call capnp.API_version) error {
	call.Results.SetVersion(1)
	return nil
}
