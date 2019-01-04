package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/gateway"
	"github.com/sahib/config"
)

func main() {
	log.SetLevel(log.DebugLevel)
	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	if err != nil {
		log.Fatalf("failed to open default config: %v", err)
	}

	cfg.SetBool("gateway.enabled", true)
	cfg.SetInt("gateway.port", 5000)
	cfg.SetBool("gateway.cert.redirect.enabled", false)

	cfg.SetBool("gateway.auth.enabled", true)
	cfg.SetString("gateway.auth.user", "ali")
	cfg.SetString("gateway.auth.pass", "ila")

	cfg.SetStrings("gateway.folders", []string{"/"})
	cfg.SetString("gateway.cert.domain", "nwzmlh4iouqikobq.myfritz.net")
	cfg.SetString("gateway.cert.certfile", "/tmp/fullchain.pem")
	cfg.SetString("gateway.cert.keyfile", "/tmp/privkey.pem")

	fs, err := catfs.NewFilesystem(
		catfs.NewMemFsBackend(),
		"/tmp/gw-standalone.db",
		"ali",
		false,
		cfg.Section("fs"),
	)

	if err != nil {
		log.Fatalf("failed to open fs: %v", err)
	}

	contents := make(map[string][]byte)
	contents["/Photos/world.png"] = []byte("Hello world")
	contents["/Photos/me.jpeg"] = []byte("me")
	contents["/Photos/somethingelse.jpeg"] = []byte("something")
	contents["/Readme.md"] = []byte("readme")
	contents["/Documents/some.pdf"] = []byte("pdf")
	contents["/Documents/Private/Nested/.hidden.txt"] = []byte("secret")

	for path, content := range contents {
		if err := fs.Stage(path, bytes.NewReader(content)); err != nil {
			log.Fatalf("failed to stage dummy data: %s %v", path, err)
		}
	}

	gw := gateway.NewGateway(fs, cfg.Section("gateway"))
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
