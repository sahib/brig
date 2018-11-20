package ipfs

import (
	ipfsversion "github.com/ipfs/go-ipfs"
)

// VersionInfo holds version info (yeah, golint)
type VersionInfo struct {
	semVer, name, rev string
}

// SemVer returns a VersionInfo string complying semantic versioning
func (v *VersionInfo) SemVer() string { return v.semVer }

// Name returns the name of the backend
func (v *VersionInfo) Name() string { return v.name }

// Rev returns the git revision of the backend
func (v *VersionInfo) Rev() string { return v.rev }

// Version returns detailed VersionInfo info as struct
func Version() *VersionInfo {
	return &VersionInfo{
		semVer: ipfsversion.CurrentVersionNumber,
		name:   "go-ipfs",
		rev:    ipfsversion.CurrentCommit,
	}
}
