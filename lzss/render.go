package lzss

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

// ErrClosed is returned when reading from (resp. writing to) or closing an
// already closed io.ReadCloser (resp. io.WriteCloser).
var ErrClosed = errors.New("lzss: reader/writer is closed")

type reader struct {
	r      io.ByteReader
	window *Window

	flags int
	match [2]byte

	buffer [maxMatchLength]byte
	n      int
	toRead []byte

	err error
}

// NewReader creates a new io.ReadCloser.
// Reads from the returned io.ReadCloser read and decompress data from r.
// If r does not also implement io.ByteReader, the io.ReadCloser may read more
// data than necessary from r.
//
// It is the caller's responsibility to call Close on the io.ReadCloser when
// done.
func NewReader(src []byte) io.ReadCloser {
	reader := new(reader)

	r := bytes.NewBuffer(src)
	reader.r = bufio.NewReader(r)

	reader.window = NewWindow()
	reader.flags = 1

	return reader
}

func (r *reader) Read(buffer []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	if len(r.toRead) > 0 {
		n := copy(buffer, r.toRead)
		r.toRead = r.toRead[n:]

		return n, nil
	}

	n, err := r.read(buffer)

	return n, err
}

// read reads and decompresses bytes from r into buffer. If buffer becomes full
// during a read, the remaining bytes will be stored into a temporary buffer and
// copied over on the next call to Read.
func (r *reader) read(buffer []byte) (int, error) {
	n := 0

	for n < len(buffer) {
		if r.flags == 1 {
			b, err := r.r.ReadByte()
			if err != nil {
				return n, err
			}

			r.flags = 0x100 | int(b)
		}

		if r.flags&1 == 1 {
			b, err := r.r.ReadByte()
			if err != nil {
				return n, err
			}

			buffer[n] = b
			n++

			r.window.WriteByte(b)
		} else {
			var err error

			r.match[0], err = r.r.ReadByte()
			if err != nil {
				return n, err
			}

			r.match[1], err = r.r.ReadByte()
			if err != nil {
				return n, err
			}

			offset := int(r.match[1])&0xf0<<4 | int(r.match[0])
			length := int(r.match[1])&0x0f + minMatchLength

			for i := 0; i < length; i++ {
				b, err := r.window.ReadByte(&offset)
				if err != nil {
					return n, err
				}

				if n < len(buffer) {
					buffer[n] = b
					n++
				} else {
					r.buffer[r.n] = b
					r.n++
				}

				r.window.WriteByte(b)
			}
		}

		r.flags >>= 1
	}

	r.toRead = r.buffer[:r.n]
	r.n = 0

	return n, nil
}

// Close closes the io.ReadCloser, but it does not close the underlying
// io.Reader.
func (r *reader) Close() error {
	if r.err != nil {
		return r.err
	}

	r.err = ErrClosed

	return nil
}
