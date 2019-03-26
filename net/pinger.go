package net

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/repo"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrPingMapClosed is returned when an operation is performed on a closed
	// ping map.
	ErrPingMapClosed = errors.New("pinger Map was already closed")

	// ErrNoSuchAddr is returned when asking for a pinger that we don't know.
	ErrNoSuchAddr = errors.New("No such addr known to ping map")
)

// PingMap remembers the times we last accessed a remote.
type PingMap struct {
	mu            sync.Mutex
	tickr         *time.Ticker
	peers         map[string]backend.Pinger
	authenticated map[string]bool
	netBk         backend.Backend
	rp            *repo.Repository
}

// NewPingMap returns a new PingMap.
func NewPingMap(rp *repo.Repository, netBk backend.Backend) *PingMap {
	pm := &PingMap{
		peers:         make(map[string]backend.Pinger),
		authenticated: make(map[string]bool),
		netBk:         netBk,
		tickr:         time.NewTicker(30 * time.Second),
		rp:            rp,
	}

	go pm.updateLoop()
	return pm
}

func (pm *PingMap) updateLoop() {
	for range pm.tickr.C {
		pm.doUpdate()
	}
}

func (pm *PingMap) doUpdate() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.peers == nil {
		return
	}

	for addr, pinger := range pm.peers {
		if pinger == nil {
			// Try to get a pinger in the background.
			// This will already update the pingmap,
			// but next time we continue in this loop.
			go pm.doUpdateSingle(addr, true)
			continue
		}

		if err := pinger.Err(); err != nil {
			// Maybe the pinger errored in between?
			log.Warningf("pinger »%s« failed: %v", addr, err)
			pinger.Close()

			// Mark this addr to be tried next time again.
			pm.peers[addr] = nil
			continue
		}

		// Reaching this point means that the pinger
		// seems to work and did not error out.
	}
}

func (pm *PingMap) doUpdateSingle(addr string, checkAuthentication bool) {
	pinger, err := pm.netBk.Ping(addr)
	if err != nil {
		if pinger != nil {
			pinger.Close()
		}
		return
	}

	// this address seems to work.
	log.Infof("pinger »%s« responding", addr)

	// this method is called in parallel:
	pm.mu.Lock()
	pm.peers[addr] = pinger
	pm.mu.Unlock()

	if !checkAuthentication {
		return
	}

	isAuthenticated := false

	// Always set the authenticated flag if we get past this point.
	defer func() {
		pm.mu.Lock()
		pm.authenticated[addr] = isAuthenticated
		pm.mu.Unlock()
	}()

	rmt, err := pm.rp.Remotes.RemoteByAddr(addr)
	if err != nil {
		log.Debugf("failed to get remote for addr »%s«: %v", addr, err)
		return
	}

	ctx := context.Background()
	conn, err := DialByAddr(ctx, addr, rmt.Fingerprint, pm.rp, pm.netBk)
	if err != nil {
		log.Infof("can ping, but not authenticated: %v", err)
		return
	}

	defer conn.Close()

	// Check if we can send them an authenticated ping message.
	// If so, we are sure they authenticated us also.
	if err := conn.Ping(); err != nil {
		return
	}

	isAuthenticated = true
}

// Sync makes sure all addresses in `addrs` are being watched.
// All currently watched addrs that are not in `addrs` are removed.
// This method does not block until all pingers have been updated.
func (pm *PingMap) Sync(addrs []string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.peers == nil {
		return ErrPingMapClosed
	}

	log.Infof("syncing ping map entries...")

	// Remember to schedule an update right after sync.
	// This will only run after Sync() due the common lock.
	go pm.doUpdate()

	// Needed for sort.Search below.
	sort.Strings(addrs)

	for _, addr := range addrs {
		if _, ok := pm.peers[addr]; !ok {
			// Just remember that we need to create a pinger for this addr.
			pm.peers[addr] = nil
		}
	}

	// Do the opposite check and see if any addrs in pm.peers
	// are not in `addrs`. If so, remove them out.
	for addr, pinger := range pm.peers {
		if pinger == nil {
			continue
		}

		idx := sort.SearchStrings(addrs, addr)
		if idx < len(addrs) && addrs[idx] == addr {
			continue
		}

		// This addr does not exist anymore.
		log.Debugf("Closing pinger for %v %v", addr, pinger)
		if err := pinger.Close(); err != nil {
			return err
		}
	}

	return nil
}

// For returns a new pinger for a certain `addr`.
func (pm *PingMap) For(addr string) (backend.Pinger, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.peers == nil {
		return nil, ErrPingMapClosed
	}

	pinger, ok := pm.peers[addr]
	if !ok {
		return nil, ErrNoSuchAddr
	}

	return pinger, nil
}

func (pm *PingMap) IsAuthenticated(addr string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	isAuthenticated, ok := pm.authenticated[addr]
	if !ok {
		return false
	}

	return isAuthenticated
}

// This is called by net handlers when we encountered an succesful connection
// opened to us.
func (pm *PingMap) markSuccesfullConnection(addr string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	go pm.doUpdateSingle(addr, false)
	pm.authenticated[addr] = true
}

// Close shuts down the ping map. Do not use afterwards.
func (pm *PingMap) Close() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Stop updateLoop as a side effect:
	pm.tickr.Stop()

	for _, pinger := range pm.peers {
		if pinger == nil {
			continue
		}

		if err := pinger.Close(); err != nil {
			return err
		}
	}

	pm.peers = nil
	return nil
}
