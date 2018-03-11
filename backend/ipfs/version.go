package ipfs

import (
	config "github.com/ipfs/go-ipfs/repo/config"
)

type version struct {
	semVer, name, rev string
}

func (v *version) SemVer() string { return v.semVer }
func (v *version) Name() string   { return v.name }
func (v *version) Rev() string    { return v.rev }

func Version() *version {
	return &version{
		semVer: config.CurrentVersionNumber,
		name:   "go-ipfs",
		rev:    config.CurrentCommit,
	}
}
