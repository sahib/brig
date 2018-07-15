package http

import (
	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	"net/http"
)

type flushfwder struct {
	cmds.ResponseEmitter
	http.Flusher
}

func NewFlushForwarder(r cmds.ResponseEmitter, f http.Flusher) ResponseEmitter {
	return flushfwder{ResponseEmitter: r, Flusher: f}
}
