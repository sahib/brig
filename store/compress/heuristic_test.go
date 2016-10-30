package compress

import (
	"bytes"
	"io"
	"testing"
)

type testCase struct {
	path         string
	reader       io.ReadSeeker
	expectedAlgo AlgorithmType
}

var (
	testCases = []testCase{
		{"1.txt", CreateAndInitByteReader(128, []byte("Small text file")), AlgoNone},
		{"2.txt", CreateAndInitByteReader(2048, []byte("Big text file")), AlgoLZ4},
		{"3.opus", CreateAndInitByteReader(128, []uint8{0x4f, 0x67, 0x67, 0x53}), AlgoNone},
		{"4.opus", CreateAndInitByteReader(2048, []uint8{0x4f, 0x67, 0x67, 0x53}), AlgoNone},
		{"5.zip", CreateAndInitByteReader(128, []uint8{0x50, 0x4b, 0x3, 0x4}), AlgoNone},
		{"6.zip", CreateAndInitByteReader(2048, []uint8{0x50, 0x4b, 0x3, 0x4}), AlgoNone},
	}
)

func CreateAndInitByteReader(len int, init []byte) io.ReadSeeker {
	slice := make([]byte, len)
	copy(slice, init)
	return bytes.NewReader(slice)
}

func TestChooseCompressAlgo(t *testing.T) {
	for _, testCase := range testCases {
		if algo, err := ChooseCompressAlgo(testCase.path, testCase.reader); err != nil {
			t.Errorf("Error: %v", err)
		} else if algo != testCase.expectedAlgo {
			t.Errorf(
				"For path '%s' expected '%s', got '%s'",
				testCase.path,
				AlgoToString[testCase.expectedAlgo],
				AlgoToString[algo],
			)
		}
	}
}
