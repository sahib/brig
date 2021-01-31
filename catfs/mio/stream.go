package mio

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/sahib/brig/catfs/mio/compress"
	"github.com/sahib/brig/catfs/mio/encrypt"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/util"
	log "github.com/sirupsen/logrus"
)

// Stream is a stream coming from the backend.
type Stream interface {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

type stream struct {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

type dumbWriterTo struct {
	r io.Reader
}

func (d dumbWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, d.r)
}

// NewOutStream creates an OutStream piping data from brig to the outside.
// `key` is used to decrypt the data. The compression algorithm is read
// from the stream header.
func NewOutStream(r io.ReadSeeker, isRaw bool, key []byte) (Stream, error) {
	s := stream{
		Reader:   r,
		Seeker:   r,
		WriterTo: dumbWriterTo{r: r},
		Closer:   ioutil.NopCloser(r),
	}

	if isRaw {
		// directly return stream.
		return s, nil
	}

	// At this point we're sure that there must be a magic number.
	// We can use it to decide what readers we should build.

	magicNumber, headerReader, err := util.PeekHeader(r, 8)
	if err != nil {
		// First read on the stream, errors will bubble up here.
		return nil, err
	}

	// make sure that the header is prefixed to the stream again:
	// compress + encrypt reader expect the magic number there.
	s.Reader = headerReader
	s.Seeker = headerReader
	s.WriterTo = dumbWriterTo{r: headerReader}

	// NOTE: Assumption here is that our own magic numbers
	//       are always 8 bytes long. Since we control it,
	//       that's reasonable.
	if len(magicNumber) != 8 {
		return nil, fmt.Errorf("bad magic number")
	}

	var isEncrypted bool

	switch mn := string(magicNumber); mn {
	case string(encrypt.MagicNumber):
		isEncrypted = true
	case string(compress.MagicNumber):
		// Not encrypted, but decompress needed.
	default:
		return nil, fmt.Errorf("unknown magic number '%s'", mn)
	}

	if isEncrypted {
		rEnc, err := encrypt.NewReader(s, key)
		if err != nil {
			return nil, err
		}

		flags, err := rEnc.Flags()
		if err != nil {
			return nil, err
		}

		s.Reader = rEnc
		s.Seeker = rEnc
		s.WriterTo = rEnc

		// The encryption header stores if we encoded the stream
		// with another stream inside (matroska like). If not,
		// we can return early.
		if flags&encrypt.FlagCompressedInside == 0 {
			return s, nil
		}
	}

	// if compression is used inside, than wrap in decompressor:
	// (s might contain decryptor or is raw stream)
	rZip := compress.NewReader(s)
	s.Reader = rZip
	s.Seeker = rZip
	s.WriterTo = rZip
	return s, nil
}

func guessCompression(path string, r io.Reader, hint *hints.Hint) (io.Reader, error) {
	// Keep the header of the file in memory, so we can do some guessing
	// of e.g. the compression algorithm we should use.
	headerReader := util.NewHeaderReader(r, 2048)
	headerBuf, err := headerReader.Peek()
	if err != nil {
		log.WithError(err).Warnf("failed to peek stream header")
		return nil, err
	}

	compressAlgo, err := compress.GuessAlgorithm(path, headerBuf)
	if err != nil {
		// NOTE: don't error out here. That just means we don't
		// guessed the perfect settings.
		log.WithError(err).
			WithField("path", path).
			Warnf("failed to guess suitable zip algorithm")
	}

	log.Debugf("guessed '%s' compression for file %s", compressAlgo, path)
	hint.CompressionAlgo = hints.CompressAlgorithmTypeToCompressionHint(compressAlgo)
	return headerReader, nil
}

// NewInStream creates a new stream that pipes data into ipfs.
// The data is read from `r`, encrypted with `key` and encoded based on the
// settings given by `hint`. `path` is only used to better guess the compression
// algorithm - if desired by `hint`. `path` can be empty.
//
// It returns a reader that will produce the encoded stream.
// If no actual encoding will be done, the second return param will be true
func NewInStream(r io.Reader, path string, key []byte, hint hints.Hint) (io.ReadCloser, bool, error) {
	var err error

	if hint.CompressionAlgo == hints.CompressionGuess {
		// replace "guess" to an actual compression algorithm.
		r, err = guessCompression(path, r, &hint)
		if err != nil {
			return nil, false, err
		}
	}

	// use a pipe to redirect `r` to encoding writers without copying:
	pr, pw := io.Pipe()

	// Writing to pw will be matched by a read on the other side.
	// If there is no read we will block.
	var w io.Writer = pw
	var closers = []io.Closer{pw}

	// Only add encryption if desired by hints:
	if hint.EncryptionAlgo != hints.EncryptionNone {
		wEnc, err := encrypt.NewWriter(w, key, hint.EncryptFlags())
		if err != nil {
			return nil, false, err
		}

		closers = append(closers, wEnc)
		w = wEnc
	}

	// Only add compression if desired or mime type is suitable:
	if hint.CompressionAlgo != hints.CompressionNone {
		wZip, err := compress.NewWriter(w, hint.CompressionAlgo.ToCompressAlgorithmType())
		if err != nil {
			return nil, false, err
		}

		closers = append(closers, wZip)
		w = wZip
	}

	// Suck the reader empty and move it to `w`.
	go func() {
		if _, err := io.Copy(w, r); err != nil {
			// Continue closing the fds; no return.
			log.WithError(err).Warnf("internal write error")
		}

		// NOTE: closers must be closed in inverse order.
		//       pipe writer should come last. Each Close()
		//       might still write out data.
		for idx := len(closers) - 1; idx >= 0; idx-- {
			if err := closers[idx].Close(); err != nil {
				log.WithError(err).Warnf("internal close error")
			}
		}
	}()

	return pr, hint.IsRaw(), nil
}

// limitedStream is a small wrapper around Stream,
// which allows truncating the stream at a certain size.
type limitedStream struct {
	stream Stream
	pos    uint64
	size   uint64
}

func (ls *limitedStream) Read(buf []byte) (int, error) {
	isEOF := false
	if ls.pos+uint64(len(buf)) >= ls.size {
		buf = buf[:ls.size-ls.pos]
		isEOF = true
	}

	n, err := ls.stream.Read(buf)
	if err != nil {
		return n, err
	}

	if isEOF {
		err = io.EOF
	}

	return n, err
}

func (ls *limitedStream) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		return ls.Seek(int64(ls.pos)+offset, io.SeekStart)
	case io.SeekEnd:
		ls.pos = 0
		return ls.Seek(int64(ls.size)+offset, io.SeekStart)
	case io.SeekStart:
		ls.pos = 0
	}

	newPos := int64(ls.pos) + offset
	if newPos < 0 {
		return -1, io.EOF
	}

	if newPos > int64(ls.size) {
		return int64(ls.size), io.EOF
	}

	ls.pos = uint64(newPos)
	return ls.stream.Seek(newPos, io.SeekStart)
}

func (ls *limitedStream) WriteTo(w io.Writer) (int64, error) {
	// We do not want to defeat the purpose of WriteTo here.
	// That's why we do the limit check in the writer part.
	return ls.stream.WriteTo(util.LimitWriter(w, int64(ls.size-ls.pos)))
}

func (ls *limitedStream) Close() error {
	return ls.stream.Close()
}

// LimitStream is like io.LimitReader, but works for mio.Stream.
// It will not allow reading/seeking after the specified size.
func LimitStream(stream Stream, size uint64) Stream {
	return &limitedStream{
		stream: stream,
		pos:    0,
		size:   size,
	}
}
