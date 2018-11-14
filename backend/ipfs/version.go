package ipfs

import (
	ipfsversion "github.com/ipfs/go-ipfs"
)

type version struct {
	semVer, name, rev string
}

func (v *version) SemVer() string { return v.semVer }
func (v *version) Name() string   { return v.name }
func (v *version) Rev() string    { return v.rev }

func Version() *version {
	return &version{
		semVer: ipfsversion.CurrentVersionNumber,
		name:   "go-ipfs",
		rev:    ipfsversion.CurrentCommit,
	}
}
