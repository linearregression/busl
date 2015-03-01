package sse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const (
	template = "id: %d\ndata: %s\n\n"
)

type encoder struct {
	reader   io.Reader     // stores the original reader
	buffered *bufio.Reader // bufio wrapper for original reader
	offset   int64         // offset for Seek purposes
}

func NewEncoder(r io.Reader) *encoder {
	return &encoder{r, bufio.NewReader(r), 0}
}

func (r *encoder) Seek(offset int64, whence int) (n int64, err error) {
	r.offset, err = r.reader.(io.ReadSeeker).Seek(offset, whence)
	return r.offset, err
}

// FIXME: this version is simplified and assumes
// that len(p) is always greater than the potential
// length of data to be read.
func (r *encoder) Read(p []byte) (n int, err error) {
	data, err := r.buffered.ReadBytes('\n')

	if n = len(data); n > 0 {
		data = format(r.offset, data)
		r.offset += int64(n)
		n = copy(p, data)
	}

	return n, err
}

func format(pos int64, msg []byte) []byte {
	id := pos + int64(len(msg))
	msg = bytes.Trim(msg, "\n")

	return []byte(fmt.Sprintf(template, id, msg))
}
