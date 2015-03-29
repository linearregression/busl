package sse

import (
	"bytes"
	"fmt"
	"io"
)

const (
	id   = "id: %d\n"
	data = "data: %s\n"
)

type encoder struct {
	reader io.Reader // stores the original reader
	offset int64     // offset for Seek purposes
}

func NewEncoder(r io.Reader) io.Reader {
	return &encoder{reader: r}
}

func (r *encoder) Seek(offset int64, whence int) (n int64, err error) {
	if seeker, ok := r.reader.(io.ReadSeeker); ok {
		r.offset, err = seeker.Seek(offset, whence)
	}
	return r.offset, err
}

// FIXME: this version is simplified and assumes
// that len(p) is always greater than the potential
// length of data to be read.
func (r *encoder) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)

	if n > 0 {
		buf := format(r.offset, p[:n])
		r.offset += int64(n)
		n = copy(p, buf)
	}

	return n, err
}

func format(pos int64, msg []byte) []byte {
	buf := bytes.NewBufferString(fmt.Sprintf(id, pos+int64(len(msg))))

	for _, line := range bytes.Split(msg, []byte{'\n'}) {
		buf.WriteString(fmt.Sprintf(data, line))
	}
	buf.Write([]byte{'\n'})

	return buf.Bytes()
}
