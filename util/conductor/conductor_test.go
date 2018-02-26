package conductor

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConductorEmpty(t *testing.T) {
	cnd := New(5*time.Second, 10)
	ticket := cnd.Exec(func(ticket uint64) error {
		return nil
	})

	_, err := cnd.Pop(ticket)
	require.True(t, IsNoDataLeft(err))
}

func TestConductorSingle(t *testing.T) {
	cnd := New(5*time.Second, 10)
	ticket := cnd.Exec(func(ticket uint64) error {
		return cnd.Push(ticket, 123)
	})

	data, err := cnd.Pop(ticket)
	require.Nil(t, err)
	require.Equal(t, data, 123)

	_, err = cnd.Pop(ticket)
	require.True(t, IsNoDataLeft(err))
}

func TestConductorMultiple(t *testing.T) {
	cnd := New(5*time.Second, 10)
	ticket := cnd.Exec(func(ticket uint64) error {
		for i := 0; i < 20; i++ {
			require.Nil(t, cnd.Push(ticket, i))
		}
		return nil
	})

	for i := 0; i < 20; i++ {
		data, err := cnd.Pop(ticket)
		require.Nil(t, err)
		require.Equal(t, data, i)
	}

	_, err := cnd.Pop(ticket)
	require.True(t, IsNoDataLeft(err))
}

func TestConductorError(t *testing.T) {
	cnd := New(5*time.Second, 10)
	ticket := cnd.Exec(func(ticket uint64) error {
		return errors.New("artificial error")
	})

	time.Sleep(10 * time.Millisecond)

	_, err := cnd.Pop(ticket)
	require.Equal(t, err.Error(), "artificial error")
}

func TestConductorTimeout(t *testing.T) {
	cnd := New(100*time.Millisecond, 10)
	ticket := cnd.Exec(func(ticket uint64) error {
		return cnd.Push(ticket, 123)
	})

	time.Sleep(10 * time.Millisecond)

	data, err := cnd.Pop(ticket)
	require.Nil(t, err)
	require.Equal(t, data, 123)

	// Check if the value is still there after a certain amount of time:
	time.Sleep(200 * time.Millisecond)
	_, err = cnd.Pop(ticket)
	require.True(t, IsNoSuchTicket(err))
}

func TestConductorTimeoutError(t *testing.T) {
	cnd := New(100*time.Millisecond, 10)
	ticket := cnd.Exec(func(ticket uint64) error {
		return errors.New("artificial error")
	})

	time.Sleep(10 * time.Millisecond)

	_, err := cnd.Pop(ticket)
	require.Equal(t, err.Error(), "artificial error")

	// Check if the value is still there after a certain amount of time:
	time.Sleep(200 * time.Millisecond)
	_, err = cnd.Pop(ticket)
	require.True(t, IsNoSuchTicket(err))
}
