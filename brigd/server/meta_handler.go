package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/brigd/capnp"
	"github.com/disorganizer/brig/repo"
	"zombiezen.com/go/capnproto2/server"
)

type metaHandler struct {
	base *base
}

func (mh *metaHandler) Quit(call capnp.Meta_quit) error {
	server.Ack(call.Options)
	log.Info("Shutting down brigd due to QUIT command")
	mh.base.QuitCh <- struct{}{}

	log.Infof("Trying to lock repository...")
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	return repo.Close(mh.base.password)
}

func (mh *metaHandler) Ping(call capnp.Meta_ping) error {
	server.Ack(call.Options)
	return call.Results.SetReply("PONG")
}

func (mh *metaHandler) Init(call capnp.Meta_init) error {
	server.Ack(call.Options)

	backendName, err := call.Params.Backend()
	if err != nil {
		return err
	}

	initFolder, err := call.Params.BasePath()
	if err != nil {
		return err
	}

	owner, err := call.Params.Owner()
	if err != nil {
		return err
	}

	return repo.Init(initFolder, owner, backendName)
}

func (mh *metaHandler) ConfigGet(call capnp.Meta_configGet) error {
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	value := repo.Config.GetString(key)
	return call.Results.SetValue(value)
}

func (mh *metaHandler) ConfigSet(call capnp.Meta_configSet) error {
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	val, err := call.Params.Value()
	if err != nil {
		return err
	}

	repo.Config.Set(key, val)
	return nil
}

func (mh *metaHandler) ConfigAll(call capnp.Meta_configAll) error {
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	all := repo.Config.AllKeys()
	seg := call.Results.Segment()

	lst, err := capnp.NewConfigPair_List(seg, int32(len(all)))
	if err != nil {
		return err
	}

	for idx, key := range all {
		pair, err := capnp.NewConfigPair(seg)
		if err != nil {
			return err
		}

		if err := pair.SetKey(key); err != nil {
			return err
		}

		if err := pair.SetVal(repo.Config.GetString(key)); err != nil {
			return err
		}

		if err := lst.Set(idx, pair); err != nil {
			return err
		}
	}

	return call.Results.SetAll(lst)
}
