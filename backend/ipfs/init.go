package ipfs

import (
	"io"
	"io/ioutil"
	"os"

	ipfsLog "gx/ipfs/QmZChCsSt8DctjceaL56Eibc29CVQq4dGKRXC5JRZ6Ppae/go-log"
	logWriter "gx/ipfs/QmZChCsSt8DctjceaL56Eibc29CVQq4dGKRXC5JRZ6Ppae/go-log/writer"
	ipfsLogging "gx/ipfs/QmcaSwFc5RBg8yCq54QURwEU4nwjfCpjbpmaAm4VbdGLKv/go-logging"

	brigLog "github.com/Sirupsen/logrus"

	ipfsconfig "gx/ipfs/QmPEpj17FDRpc7K1aArKZp3RsHtzRMKykeK9GVgn4WQGPR/go-ipfs-config"

	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

// Init creates an initialized .ipfs directory in the directory `path`.
// The generated RSA key will have `keySize` bits.
func Init(path string, keySize int) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}

	// init, but discard the log messages about generating a key.
	cfg, err := ipfsconfig.Init(ioutil.Discard, keySize)
	if err != nil {
		return err
	}

	// Init the actual data store.
	if err := fsrepo.Init(path, cfg); err != nil {
		return err
	}

	return nil
}

// ForwardLog routes all ipfs logs to a file provided by brig.
func ForwardLog(w io.Writer) {
	logWriter.Configure(logWriter.Output(w))
	ipfsLogging.SetLevel(ipfsLogging.WARNING, "*")

	if err := ipfsLog.SetLogLevel("*", "warning"); err != nil {
		brigLog.Errorf("failed to set ipfs log level: %v", err)
	}
}
