package compress

import (
	"testing"
)

type testCase struct {
	path         string
	size         uint64
	header       []byte
	expectedAlgo AlgorithmType
}

var (
	testCases = []testCase{
		{"1.txt", 128, []byte("Small text file"), AlgoNone},
		{"2.txt", 2048, []byte("Big text file"), AlgoLZ4},
		{"3.opus", 128, []byte{0x4f, 0x67, 0x67, 0x53}, AlgoNone},
		{"4.opus", 2048, []byte{0x4f, 0x67, 0x67, 0x53}, AlgoNone},
		{"5.zip", 128, []byte{0x50, 0x4b, 0x3, 0x4}, AlgoNone},
		{"6.zip", 2048, []byte{0x50, 0x4b, 0x3, 0x4}, AlgoNone},
	}
)

func TestChooseCompressAlgo(t *testing.T) {
	for _, tc := range testCases {
		if algo, err := ChooseCompressAlgo(tc.path, tc.size, tc.header); err != nil {
			t.Errorf("Error: %v", err)
		} else if algo != tc.expectedAlgo {
			t.Errorf(
				"For path '%s' expected '%s', got '%s'",
				tc.path,
				AlgoToString[tc.expectedAlgo],
				AlgoToString[algo],
			)
		}
	}
}
