package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
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
