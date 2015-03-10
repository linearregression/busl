package sse

import (
	. "gopkg.in/check.v1"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

type table struct {
	offset int64
	input  string
	output string
}

var (
	testdata = []table{
		{0, "hello", "id: 5\ndata: hello\n\n"},
		{0, "hello\n", "id: 6\ndata: hello\ndata: \n\n"},
		{0, "hello\nworld", "id: 11\ndata: hello\ndata: world\n\n"},
		{0, "hello\nworld\n", "id: 12\ndata: hello\ndata: world\ndata: \n\n"},
		{1, "hello\nworld\n", "id: 12\ndata: ello\ndata: world\ndata: \n\n"},
		{6, "hello\nworld\n", "id: 12\ndata: world\ndata: \n\n"},
		{11, "hello\nworld\n", "id: 12\ndata: \ndata: \n\n"},
		{12, "hello\nworld\n", ""},
	}
)

func Test(t *testing.T) { TestingT(t) }

type SseSuite struct{}

var _ = Suite(&SseSuite{})

func (s *SseSuite) TestNoNewline(c *C) {
	for _, t := range testdata {
		r := strings.NewReader(t.input)
		enc := NewEncoder(r)
		enc.(io.Seeker).Seek(t.offset, 0)
		c.Assert(readstring(enc), Equals, t.output)
	}
}

func readstring(r io.Reader) string {
	buf, _ := ioutil.ReadAll(r)
	return string(buf)
}
