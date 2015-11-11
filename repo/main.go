package main

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

type Repository interface {
	Open()
	Close()
	Lock()
	Unlock()
}

type FsRepository struct {
	Jid      string // name@domain.tld
	Password string // sha-x representation
	Folder   string // filesystem foldername repo is in
	Uid      string

	ConfigPath string

	// Crypto
	PublicKey  string
	PrivateKey string
	AesKey     string
	OtrKey     string
}

// Interface methods
func (r *FsRepository) Open() {
	fmt.Println("Opening repository.")
}

func (r *FsRepository) Close() {
	fmt.Println("Closing repository.")
}

func (r *FsRepository) Lock() {
	fmt.Println("Locking repository.")
}

func (r *FsRepository) Unlock() {
	fmt.Println("Unlocking repository.")
}

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
	repoUid := uuid.NewRandom()
	repo := FsRepository{
		Jid:        jid,
		Password:   pass,
		Folder:     absFolderPath,
		Uid:        repoUid.String(),
		ConfigPath: path.Join(folder, ".brig", "config"),
	}
	return &repo, nil
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
	} else {
		return createIPFS(path.Join(absFolderPath, ".ipfs"))
	}

	return nil
}

func CloneFsRepository() *Repository {
	return nil
}

func createIPFS(ipfsfolder string) error {

	// https://ipfs.io/ipfs/
	// QmTkzDwWqPbnAh5YiV5VwcTLnGdwSNsNTn2aDxdXBFca7D/
	// example#/ipfs/QmQwAP9vFjbCtKvD8RkJdCvPHqLQjZfW7Mqbbqx18zd8j7/api/service/readme.md
	// dosn't work the way expected

	// Is this the right way?
	os.Setenv("IPFS_PATH", ipfsfolder)
	cmd := exec.Command("ipfs", "init")
	err := cmd.Start()
	if err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func main() {
	_, err := NewFsRepository(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
