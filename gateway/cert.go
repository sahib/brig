package gateway

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sahib/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

func getTLSConfig(cfg *config.Config) (*tls.Config, error) {
	// In case we don't have a cert or it is outdated, we should
	// get it in the background. This sets the cert.certfile
	// and cert.keyfile as side effect on success.
	if err := updateCert(cfg); err != nil {
		log.Warningf("failed to get certificate: %v", err)
	}

	certPath := cfg.String("cert.certfile")
	keyPath := cfg.String("cert.keyfile")
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}

		// PCI DSS 3.2.1. demands offering TLS >= 1.1:
		return &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS11,
			PreferServerCipherSuites: true,
		}, nil
	}

	return nil, nil
}

func encodeECDSAKey(w io.Writer, key *ecdsa.PrivateKey) error {
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	pb := &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	return pem.Encode(w, pb)
}

func certToPubPrivKeyPair(tlscert *tls.Certificate) ([]byte, []byte, error) {
	// contains PEM-encoded data
	var privBuf bytes.Buffer
	var pubBuf bytes.Buffer

	// private
	switch key := tlscert.PrivateKey.(type) {
	case *ecdsa.PrivateKey:
		if err := encodeECDSAKey(&privBuf, key); err != nil {
			return nil, nil, err
		}
	case *rsa.PrivateKey:
		b := x509.MarshalPKCS1PrivateKey(key)
		pb := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: b}
		if err := pem.Encode(&privBuf, pb); err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, errors.New("acme/autocert: unknown private key type")
	}

	// public
	for _, b := range tlscert.Certificate {
		pb := &pem.Block{Type: "CERTIFICATE", Bytes: b}
		if err := pem.Encode(&pubBuf, pb); err != nil {
			return nil, nil, err
		}
	}

	return privBuf.Bytes(), pubBuf.Bytes(), nil
}

// UserCacheDir is the same as os.UserCacheDir from go1.11,
// but taken from the standard library. This way it also works
// for go1.9 and go1.10.
//
// This method should be replaced by os.UserCacheDir by go1.13.
func UserCacheDir() (string, error) {
	var dir string

	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}

	case "darwin":
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}
		dir += "/Library/Caches"

	case "plan9":
		dir = os.Getenv("home")
		if dir == "" {
			return "", errors.New("$home is not defined")
		}
		dir += "/lib/cache"

	default: // Unix
		dir = os.Getenv("XDG_CACHE_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				return "", errors.New("neither $XDG_CACHE_HOME nor $HOME are defined")
			}
			dir += "/.cache"
		}
	}

	return dir, nil
}

// FetchTLSCertificate will use the ACME protocol and LetsEncrypt to
// download a certificate to the user's cache dir automatically.
// This either needs rights to bind :80 (i.e. sudo) or the right capabilities
// (i.e. sudo setcap CAP_NET_BIND_SERVICE=+ep ~/go/bin/brig)
func FetchTLSCertificate(domain string, cacheDir string) (string, string, error) {
	if cacheDir == "" {
		var err error
		cacheDir, err = UserCacheDir()
		if err != nil {
			return "", "", err
		}
	}

	cacheDir = filepath.Join(cacheDir, "brig")
	if err := os.MkdirAll(cacheDir, 0750); err != nil {
		return "", "", err
	}

	privPath := filepath.Join(cacheDir, fmt.Sprintf("%s_key.pem", domain))
	pubPath := filepath.Join(cacheDir, fmt.Sprintf("%s_cert.pem", domain))

	// Try to bind to port :80 as early as possible.
	lst, err := net.Listen("tcp", ":80") // #nosec
	if err != nil {
		// This most likely failed if we do not have access to port 80.
		// Try to use the (possibly outdated) cert from the cache, if there.
		_, privPathErr := os.Stat(privPath)
		_, pubPathErr := os.Stat(pubPath)
		if privPathErr == nil && pubPathErr == nil {
			log.Debugf("could not update certificate, but found cached one.")
			return privPath, pubPath, nil
		}

		return "", "", err
	}

	mgr := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache(cacheDir),
	}

	go func() {
		defer lst.Close()

		srv := &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      mgr.HTTPHandler(nil),
		}

		if err := srv.Serve(lst); err != nil {
			log.Fatalf("httpSrv.ListenAndServe() failed with %s", err)
		}
	}()

	// Silly way to make sure the :80 server is running already.
	time.Sleep(50 * time.Millisecond)

	cert, err := mgr.GetCertificate(&tls.ClientHelloInfo{
		ServerName: domain,
	})

	if err != nil {
		return "", "", err
	}

	privData, pubData, err := certToPubPrivKeyPair(cert)
	if err != nil {
		return "", "", err
	}

	// HACK: This function gets also executed by "brig gateway cert",
	// which tends to be run as root. We store the certificate in ~/.cache/brig
	// anyways, so even it is owned by root, it should be readable by other users.
	perms := os.FileMode(0600)
	if os.Geteuid() == 0 {
		perms = os.FileMode(0640) // #nosec
	}

	fmt.Println("WRITING", privPath, pubPath)
	if err = ioutil.WriteFile(privPath, privData, perms); err != nil {
		return "", "", err
	}

	if err = ioutil.WriteFile(pubPath, pubData, perms); err != nil {
		return "", "", err
	}

	return privPath, pubPath, nil
}

func updateCert(cfg *config.Config) error {
	domain := cfg.String("cert.domain")
	if domain == "" {
		log.Debugf("note: no domain set in config, cannot update certificate")
		return nil
	}

	privPath, pubPath, err := FetchTLSCertificate(domain, "")
	if err != nil {
		return err
	}

	cfg.SetString("cert.keyfile", privPath)
	cfg.SetString("cert.certfile", pubPath)
	return nil
}
