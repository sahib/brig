package httpipfs

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
func (n *Node) Version() *VersionInfo {
	v, rev, err := n.sh.Version()
	if err != nil {
		return nil
	}

	return &VersionInfo{
		semVer: v,
		name:   "go-ipfs",
		rev:    rev,
	}
}
