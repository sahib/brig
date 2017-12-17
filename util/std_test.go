package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/sahib/brig/util/testutil"
)

func TestClamp(t *testing.T) {
	if Clamp(-1, 0, 1) != 0 {
		t.Errorf("Clamp: -1 is not in [0, 1]")
	}

	if Clamp(+1, 0, 1) != 1 {
		t.Errorf("Clamp: +1 should be [0, 1]")
	}

	if Clamp(0, 0, 1) != 0 {
		t.Errorf("Clamp: 0 should be [0, 1]")
	}

	if Clamp(+2, 0, 1) != 1 {
		t.Errorf("Clamp: 2 was not cut")
	}
}

func TestSizeAcc(t *testing.T) {
	N := 20
	data := []byte("Hello World, how are you today?")

	sizeAcc := &SizeAccumulator{}
	buffers := []*bytes.Buffer{}

	for i := 0; i < N; i++ {
		buf := bytes.NewBuffer(data)
		buffers = append(buffers, buf)
	}

	wg := &sync.WaitGroup{}
	wg.Add(N)

	for i := 0; i < N; i++ {
		go func(buf *bytes.Buffer) {
			for j := 0; j < len(data); j++ {
				miniBuf := []byte{0}
				buf.Read(miniBuf)
				if _, err := sizeAcc.Write(miniBuf); err != nil {
					t.Errorf("write(sizeAcc, miniBuf) failed: %v", err)
				}
			}

			wg.Done()
		}(buffers[i])
	}

	wg.Wait()
	if int(sizeAcc.Size()) != N*len(data) {
		t.Errorf("SizeAccumulator: Sizes got dropped, race condition?")
		t.Errorf(
			"Should be %v x %v = %v; was %v",
			len(data), N, len(data)*N, sizeAcc.Size(),
		)
	}
}

func TestTouch(t *testing.T) {
	// Test for fd leakage:
	N := 4097

	baseDir := filepath.Join(os.TempDir(), "touch-test")
	if err := os.Mkdir(baseDir, 0777); err != nil {
		t.Errorf("touch-test: Could not create temp dir: %v", err)
		return
	}

	defer func() {
		if err := os.RemoveAll(baseDir); err != nil {
			t.Errorf("touch-test: Could not remove temp-dir: %v", err)
		}
	}()

	for i := 0; i < N; i++ {
		touchPath := filepath.Join(baseDir, fmt.Sprintf("%d", i))
		if err := Touch(touchPath); err != nil {
			t.Errorf("touch-test: Touch() failed: %v", err)
			return
		}

		if _, err := os.Stat(touchPath); os.IsNotExist(err) {
			t.Errorf("touch-test: `%v` does not exist after Touch()", touchPath)
			return
		}
	}
}

type slowWriter struct{}

func (w slowWriter) Write(buf []byte) (int, error) {
	time.Sleep(500 * time.Millisecond)
	return 0, nil
}

func TestTimeoutWriter(t *testing.T) {
	fast := NewTimeoutWriter(&bytes.Buffer{}, 500*time.Millisecond)
	beforeFast := time.Now()
	fast.Write([]byte("Hello World"))
	fastTook := time.Since(beforeFast)

	if fastTook > 50*time.Millisecond {
		t.Errorf("TimeoutWriter did wait too long.")
		return
	}

	beforeSlow := time.Now()
	slow := NewTimeoutWriter(slowWriter{}, 250*time.Millisecond)
	slow.Write([]byte("Hello World"))
	slowTook := time.Since(beforeSlow)

	if slowTook > 300*time.Millisecond {
		t.Errorf("TimeoutWriter did not kill write fast enough.")
		return
	}

	if slowTook < 200*time.Millisecond {
		t.Errorf("TimeoutWriter did return too fast.")
		return
	}
}

func ExampleSizeAccumulator() {
	s := &SizeAccumulator{}
	teeR := io.TeeReader(bytes.NewReader([]byte("Hello, ")), s)
	io.Copy(os.Stdout, teeR)
	fmt.Printf("wrote %d bytes to stdout\n", s.Size())
	// Output: Hello, wrote 7 bytes to stdout
}

func TestLimitWriterSimple(t *testing.T) {
	tcs := []struct {
		limit     int64
		dummySize int64
		writes    int
		name      string
	}{
		{1024, 512, 3, "basic"},
		{1024, 512, 2, "exact"},
		{1022, 511, 2, "off-by-two"},
		{1023, 1024, 1, "off-mimus-one"},
		{1024, 1025, 1, "off-plus-one"},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			w := LimitWriter(outBuf, tc.limit)

			dummy := testutil.CreateDummyBuf(tc.dummySize)
			expected := make([]byte, 0)
			for i := 0; i < tc.writes; i++ {
				w.Write(dummy)
				expected = append(expected, dummy...)
			}

			expected = expected[:tc.limit]

			if outBuf.Len() != int(tc.limit) {
				t.Fatalf(
					"Length differs (got %d, want %d)",
					outBuf.Len(),
					tc.limit,
				)
			}

			if !bytes.Equal(expected, outBuf.Bytes()) {
				t.Fatalf("Data differs")
			}
		})
	}
}
