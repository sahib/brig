package ipfsutil

import (
	"io/ioutil"
	"os"

	ipfsconfig "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

// InitRepo creates an initialized .ipfs directory in the directory `path`.
// The generated RSA key will have `keySize` bits.
func InitRepo(path string, keySize int) (string, error) {
	if err := os.MkdirAll(path, 0744); err != nil {
		return "", err
	}

	cfg, err := ipfsconfig.Init(ioutil.Discard, keySize)
	if err != nil {
		return "", err
	}

	if err := fsrepo.Init(path, cfg); err != nil {
		return "", err
	}

	return path, nil
}
