package mock

import (
	"fmt"
	"time"
)

type dummyPinger struct {
	lastSeen  time.Time
	roundtrip time.Duration
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
	return nil
}

func pingerByName(addr string) (*dummyPinger, error) {
	switch addr {
	case "alice-addr":
		return &dummyPinger{
			roundtrip: time.Duration(0),
			lastSeen:  time.Now(),
			err:       nil,
		}, nil
	case "vincent-addr":
		return nil, fmt.Errorf("vincent is offline")
	case "bob-addr":
		return &dummyPinger{
			roundtrip: 42 * time.Millisecond,
			lastSeen:  time.Now().Add(-1 * time.Minute),
		}, nil
	case "charlie-addr-right":
		return &dummyPinger{
			roundtrip: time.Duration(23),
			lastSeen:  time.Now().Add(-5 * time.Minute),
		}, nil
	case "charlie-addr-wrong":
		time.Sleep(5 * time.Second)
		return &dummyPinger{
			roundtrip: time.Duration(23),
			lastSeen:  time.Now().Add(-15 * time.Minute),
			err:       fmt.Errorf("Connection lost"),
		}, nil
	default:
		return nil, fmt.Errorf("Unknown mock peer addr: %v", addr)
	}
}
