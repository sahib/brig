package repo

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
	"os"
	"path"
	"path/filepath"
)

func generateKeypair(name, comment, email string, cfg *packet.Config) (*openpgp.Entity, error) {
	fmt.Println("Generating keypair...")
	keypair, err := openpgp.NewEntity(name, comment, email, cfg)
	if err != nil {
		return nil, err
	}
	fmt.Println("...done.")
	return keypair, nil
}

func saveKeypair(path string, keypair *openpgp.Entity) error {
	// TODO: Save key armored with a passphrase - currently not supported by
	// golang openpgp library
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	keypair.SerializePrivate(f, nil) // TODO: packet.Config instead of nil?
	return nil
}

func loadKeypair(path string) (*openpgp.Entity, error) {
	// TODO: Save key armored keypair
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	enityList, err := openpgp.ReadKeyRing(f)
	if err != nil {
		return nil, err
	}
	if len(enityList) > 0 {
		return enityList[0], nil
	}
	// should acually never be reached
	return nil, err
}

// Repository interface for brig repository types
type Repository interface {
	Open()
	Close()
	Lock()
	Unlock()
}

// FsRepository represents data a brig repository consists of
type FsRepository struct {
	Jid      string // name@domain.tld
	Password string // sha-x representation
	Folder   string // filesystem foldername repo is in
	UID      string

	ConfigPath string

	// Crypto
	GPGKeypair *openpgp.Entity
	AesKey     string
	OtrKey     string
}

// Interface methods

// Open a encrypted repository
func (r *FsRepository) Open() {
	fmt.Println("Opening repository.")
}

// Close a open repository
func (r *FsRepository) Close() {
	fmt.Println("Closing repository.")
}

// Lock a repository to be read only
func (r *FsRepository) Lock() {
	fmt.Println("Locking repository.")
}

// Unlock a repository to be writeable
func (r *FsRepository) Unlock() {
	fmt.Println("Unlocking repository.")
}

// NewFsRepository creates a new repository at filesystem level
// and returns a Repository interface
func NewFsRepository(jid, pass, folder string) (Repository, error) {

	absFolderPath, err := filepath.Abs(folder)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(absFolderPath); os.IsExist(err) {
		return nil, err
	}

	if err := createRepositoryTree(absFolderPath); err != nil {
		return nil, err
	}

	repoUID := uuid.NewRandom()
	repo := FsRepository{
		Jid:        jid,
		Password:   pass,
		Folder:     absFolderPath,
		UID:        repoUID.String(),
		ConfigPath: path.Join(folder, ".brig", "config"),
	}
	return &repo, nil
}

// CloneFsRepository clones a brig repository in a git like way
func CloneFsRepository() *Repository {
	return nil
}

func createRepositoryTree(absFolderPath string) error {
	if err := os.Mkdir(absFolderPath, 0755); err != nil {
		return err
	}

	if err := os.Mkdir(path.Join(absFolderPath, ".brig"), 0755); err != nil {
		return err
	}

	if err := os.Mkdir(path.Join(absFolderPath, ".ipfs"), 0755); err != nil {
		return err
	}

	return createIPFS(path.Join(absFolderPath, ".ipfs"))
}

func createIPFS(ipfsRootPath string) error {
	cfg, err := config.Init(os.Stdout, 2048)
	if err != nil {
		return err
	}

	if err := fsrepo.Init(ipfsRootPath, cfg); err != nil {
		return err
	}

	return nil
}

//func main() {
//	//_, err := NewFsRepository(os.Args[1], os.Args[2], os.Args[3])
//	//if err != nil {
//	//	fmt.Println(err)
//	//	os.Exit(3)
//	//}
//	e, err := loadKeypair("catz")
//	if err != nil {
//		fmt.Println(err)
//	}
//	for _, subkey := range e.Subkeys {
//		if subkey.PrivateKey != nil {
//			fmt.Printf("Privater Schluessel % x\n", subkey.PrivateKey.Fingerprint)
//		}
//		if subkey.PublicKey != nil {
//			fmt.Printf("Public Schluessel % x\n", subkey.PublicKey.Fingerprint)
//		}
//	}
//}
