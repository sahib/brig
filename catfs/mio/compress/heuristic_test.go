package compress

import (
	"testing"

	"github.com/sahib/brig/util/testutil"
)

type testCase struct {
	path         string
	header       []byte
	expectedAlgo AlgorithmType
}

var (
	testCases = []testCase{
		{
			"1.txt",
			testutil.CreateDummyBuf(HeaderSizeThreshold - 1),
			AlgoNone,
		}, {
			"2.txt",
			testutil.CreateDummyBuf(HeaderSizeThreshold),
			AlgoLZ4,
		}, {
			"3.opus",
			append(
				[]byte{0x4f, 0x67, 0x67, 0x53},
				testutil.CreateDummyBuf(HeaderSizeThreshold)...,
			),
			AlgoNone,
		}, {
			"4.zip",
			append(
				[]byte{0x50, 0x4b, 0x3, 0x4},
				testutil.CreateDummyBuf(HeaderSizeThreshold)...,
			),
			AlgoNone,
		},
	}
)

func TestChooseCompressAlgo(t *testing.T) {
	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			if algo, err := GuessAlgorithm(tc.path, tc.header); err != nil {
				t.Errorf("Error: %v", err)
			} else if algo != tc.expectedAlgo {
				t.Errorf(
					"For path '%s' expected '%s', got '%s'",
					tc.path,
					AlgoToString[tc.expectedAlgo],
					AlgoToString[algo],
				)
			}
		})
	}
}
