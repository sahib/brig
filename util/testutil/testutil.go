package testutil

import "io/ioutil"

// CreateDummyBuf creates a byte slice that is `size` big.
// It's filled with the repeating numbers [0...255].
func CreateDummyBuf(size int64) []byte {
	buf := make([]byte, size)

	for i := int64(0); i < size; i++ {
		// Be evil and stripe the data:
		buf[i] = byte(i % 255)
	}

	return buf
}

// CreateFile creates a temporary file in the systems tmp-folder.
// The file will be `size` bytes big, filled with content from CreateDummyBuf.
func CreateFile(size int64) string {
	fd, err := ioutil.TempFile("", "brig_test")
	if err != nil {
		panic("Cannot create temp file")
	}

	defer fd.Close()

	blockSize := int64(1 * 1024 * 1024)
	buf := CreateDummyBuf(blockSize)

	for size > 0 {
		take := size
		if size > int64(len(buf)) {
			take = int64(len(buf))
		}

		_, err := fd.Write(buf[:take])
		if err != nil {
			panic(err)
		}

		size -= blockSize
	}

	return fd.Name()
}
