package gateway

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/config"
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

		return &tls.Config{
			Certificates: []tls.Certificate{cert},
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

// FetchTLSCertificate will use the ACME protocol and LetsEncrypt to
// download a certificate to the user's cache dir automatically.
// This either needs rights to bind :80 (i.e. sudo) or the right capabilities
// (i.e. sudo setcap CAP_NET_BIND_SERVICE=+ep ~/go/bin/brig)
func FetchTLSCertificate(domain string) (string, string, error) {
	// Try to bind to port :80 as early as possible.
	lst, err := net.Listen("tcp", ":80")
	if err != nil {
		return "", "", err
	}

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		lst.Close()
		return "", "", err
	}

	cacheDir := filepath.Join(userCacheDir, "brig")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		lst.Close()
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

	privPath := filepath.Join(cacheDir, "key.pem")
	if err = ioutil.WriteFile(privPath, privData, 0600); err != nil {
		return "", "", err
	}

	pubPath := filepath.Join(cacheDir, "cert.pem")
	if err = ioutil.WriteFile(pubPath, pubData, 0600); err != nil {
		return "", "", err
	}

	return privPath, pubPath, nil
}

func updateCert(cfg *config.Config) error {
	domain := cfg.String("cert.domain")
	if domain == "" {
		log.Debugf("note: no domain set in config, cannot get certificate")
		return nil
	}

	privPath, pubPath, err := FetchTLSCertificate(domain)
	if err != nil {
		return err
	}

	cfg.SetString("cert.keyfile", privPath)
	cfg.SetString("cert.certfile", pubPath)
	return nil
}
