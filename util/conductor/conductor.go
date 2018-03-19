// Package conductor is a small helper to execute work heavy operations
// in the backgrounds that deliver partial results ("result streaming").
package conductor

import (
	"fmt"
	"sync"
	"time"
)

// ErrNoSuchTicket will be returned by Push() or Pop() if there is no such ticket
// with this id. This will be also returned when accessing a previosuly expired ticket.
type ErrNoSuchTicket struct {
	ticket uint64
}

func (ens ErrNoSuchTicket) Error() string {
	return fmt.Sprintf("no data associated with `%d` (or it timed out)", ens.ticket)
}

// IsNoSuchTicket returns true if `err` is a ErrNoSuchTicket error.
func IsNoSuchTicket(err error) bool {
	_, ok := err.(ErrNoSuchTicket)
	return ok
}

// ErrNoDataLeft is returned by Pop() if there is no data left to be returned.
// This will only happen if the exec func returned and there are no cached results.
type ErrNoDataLeft struct {
	ticket uint64
}

func (end ErrNoDataLeft) Error() string {
	return fmt.Sprintf("No data left for ticket `%d`", end.ticket)
}

// IsNoDataLeft returns true if `err` is a ErrNoDataLeft error.
func IsNoDataLeft(err error) bool {
	_, ok := err.(ErrNoDataLeft)
	return ok
}

// Conductor (as in train conductor) yields tickets for jobs.
// There are three operations:
// - Exec: Execute heavy work in the background, caller get's a ticket.
// - Push: Report partial results of the heavy work.
// - Pull: Fetch a partial result from the caller side via the ticket.
//
// Tickets are only job ids, i.e. plain integers.
// The concept is similar to futures, but the result will clean up
// themselves after a certain timeout.
type Conductor struct {
	mu          sync.Mutex
	timeout     time.Duration
	maxData     int
	ticketCount uint64
	tickets     map[uint64]chan interface{}
	errors      map[uint64]error
}

// New creates a new conductor that will expire unfetched results
// after `timeout` and will hold at max `maxData` partial results in memory.
func New(timeout time.Duration, maxData int) *Conductor {
	return &Conductor{
		timeout: timeout,
		maxData: maxData,
		tickets: make(map[uint64]chan interface{}),
		errors:  make(map[uint64]error),
	}
}

// Exec executes `fn` in the background. `fn` gets the current ticket and can
// use it to push partial results using Push(). Exec will return a ticket number
// that can be used by the caller to Pull() partial results until ErrNoDataLeft
// is returned.
func (cd *Conductor) Exec(fn func(ticket uint64) error) uint64 {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cd.ticketCount++
	ticket := cd.ticketCount
	cd.tickets[ticket] = make(chan interface{}, cd.maxData)

	go func() {
		if err := fn(ticket); err != nil {
			cd.mu.Lock()
			cd.errors[ticket] = err
			cd.mu.Unlock()
		}

		cd.mu.Lock()
		if ticketCh, ok := cd.tickets[ticket]; ok {
			close(ticketCh)
		}
		cd.mu.Unlock()

		time.Sleep(cd.timeout)

		cd.mu.Lock()
		delete(cd.tickets, ticket)
		delete(cd.errors, ticket)
		cd.mu.Unlock()
	}()

	return ticket
}

// Push records a new partial result under a specific `ticket`.
// Push will block after pushing more than `maxData` items.
func (cd *Conductor) Push(ticket uint64, data interface{}) error {
	cd.mu.Lock()

	if err, ok := cd.errors[ticket]; ok && err != nil {
		cd.mu.Unlock()
		return err
	}

	if _, ok := cd.tickets[ticket]; !ok {
		cd.mu.Unlock()
		return fmt.Errorf("push outside exec fn is forbidden")
	}
	cd.mu.Unlock()

	cd.tickets[ticket] <- data
	return nil
}

// This is it's own function to make use of defer possible.
func (cd *Conductor) fetchTicketCh(ticket uint64) (chan interface{}, error) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if err, ok := cd.errors[ticket]; ok && err != nil {
		return nil, err
	}

	ticketCh, ok := cd.tickets[ticket]
	if !ok {
		return nil, ErrNoSuchTicket{ticket}
	}

	return ticketCh, nil
}

// Pop gets the first (FIFO) partial result associated with `ticket`.
// It will return ErrNoDataLeft if there won't be any more values.
// It will return ErrNoSuchTicket if you passed an invalid ticket.
func (cd *Conductor) Pop(ticket uint64) (interface{}, error) {
	ticketCh, err := cd.fetchTicketCh(ticket)
	if err != nil {
		return nil, err
	}

	// Wait until we get results:
	timer := time.NewTimer(cd.timeout)
	select {
	case data, ok := <-ticketCh:
		if !ok {
			return nil, ErrNoDataLeft{ticket}
		}

		return data, nil
	case <-timer.C:
		return nil, fmt.Errorf("pop took too long (%v)", cd.timeout)
	}
}
