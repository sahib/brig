package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/gateway"
	"github.com/sahib/config"
)

const dbPath = "/tmp/gw-runner"

func main() {
	log.SetLevel(log.DebugLevel)
	if err := os.MkdirAll(dbPath, 0700); err != nil {
		log.Fatalf("failed to create dir %s: %v", dbPath, err)
	}

	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	if err != nil {
		log.Fatalf("failed to open default config: %v", err)
	}

	cfg.SetBool("gateway.enabled", true)
	cfg.SetInt("gateway.port", 5000)
	cfg.SetBool("gateway.cert.redirect.enabled", false)

	cfg.SetBool("gateway.auth.enabled", true)
	cfg.SetString("gateway.auth.user", "admin")
	cfg.SetString("gateway.auth.pass", "password")

	cfg.SetStrings("gateway.folders", []string{"/"})
	cfg.SetString("gateway.cert.domain", "nwzmlh4iouqikobq.myfritz.net")
	cfg.SetString("gateway.cert.certfile", "/tmp/fullchain.pem")
	cfg.SetString("gateway.cert.keyfile", "/tmp/privkey.pem")

	bk, err := NewTmpFsBackend(filepath.Join(dbPath, "backend"))
	if err != nil {
		log.Fatalf("failed to open backend: %v", err)
	}

	fsPath := filepath.Join(dbPath, "metadata")
	fs, err := catfs.NewFilesystem(bk, fsPath, "ali", false, cfg.Section("fs"))
	if err != nil {
		log.Fatalf("failed to open fs: %v", err)
	}

	for _, root := range os.Args[1:] {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			fd, err := os.Open(path)
			if err != nil {
				log.Fatalf("failed to open: %v", err)
			}

			if err := fs.Stage(path[len(root):], fd); err != nil {
				log.Fatalf("failed to stage: %s: %v", path, err)
			}

			return fd.Close()
		})

		if err != nil {
			log.Fatalf("walk failed: %v", err)
		}
	}

	gw := gateway.NewGateway(fs, cfg.Section("gateway"), nil)
	gw.Start()

	defer func() {
		if err := gw.Stop(); err != nil {
			log.Warningf("failed to stop properly: %v", err)
		}
	}()

	// Block until hitting Ctrl-C
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	fmt.Println("Hit Ctrl-C to interrupt.")
	<-ch
	fmt.Println("Interrupted. Shutting down.")
}
