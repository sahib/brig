package mock

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

type dummyPinger struct {
	lastSeen  time.Time
	roundtrip time.Duration
	quitCh    chan bool
	err       error
}

func (dp *dummyPinger) LastSeen() time.Time {
	return dp.lastSeen
}

func (dp *dummyPinger) Roundtrip() time.Duration {
	return dp.roundtrip
}

func (dp *dummyPinger) Err() error {
	return dp.err
}

func (dp *dummyPinger) Close() error {
	dp.quitCh <- true
	return nil
}

func (dp *dummyPinger) ping(port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		dp.err = fmt.Errorf("no route: %v", err)
		return
	}

	conn.Close()

	// Simulate a successful ping:
	jitter := time.Duration(rand.Intn(5000)) * time.Nanosecond
	dp.roundtrip = 20*time.Millisecond + jitter
	dp.lastSeen = time.Now()
	return
}

func pingerByName(addr string) (*dummyPinger, error) {
	port, err := getPortFromAddr(addr)
	if err != nil {
		return nil, err
	}

	dp := &dummyPinger{
		quitCh: make(chan bool, 10),
	}

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				dp.ping(port)
			case <-dp.quitCh:
				return
			}
		}
	}()

	return dp, nil
}
