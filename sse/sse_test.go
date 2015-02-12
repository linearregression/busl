package sse

import (
	. "gopkg.in/check.v1"
	"testing"
)

var (
	input = []string{
		"hello",
		"world",
		"!",
		"hello\nworld",
		"hello\n\nworld",
	}

	output = []string{
		"id: 5\ndata: hello\n\n",
		"id: 10\ndata: world\n\n",
		"id: 11\ndata: !\n\n",
		"id: 22\ndata: hello\ndata: world\n\n",
		"id: 34\ndata: hello\ndata: \ndata: world\n\n",
	}
)

func Test(t *testing.T) { TestingT(t) }

type SseSuite struct{}

var _ = Suite(&SseSuite{})

func channels() (chan []byte, chan []byte) {
	in := make(chan []byte, 10)
	out := make(chan []byte, 10)

	return in, out
}

func (s *SseSuite) TestInOut(c *C) {
	in, out := channels()
	go Transform(0, in, out)
	defer close(in)

	for i, v := range output {
		in <- []byte(input[i])
		res := <-out

		if string(res) != v {
			c.Assert(string(res), Equals, v)
		}
	}
}

func (s *SseSuite) TestEvenOffset(c *C) {
	in, out := channels()
	go Transform(5, in, out)
	defer close(in)

	for i, v := range output[1:] {
		in <- []byte(input[i+1])
		res := <-out

		c.Assert(string(res), Equals, v)
	}
}

func (s *SseSuite) TestAwkwardOffsetOriginalChunks(c *C) {
	in, out := channels()
	go Transform(3, in, out)
	defer close(in)

	in <- []byte(input[0][3:])
	res := <-out

	c.Assert(string(res), Equals, string(format(3, []byte("lo"))))
}

func (s *SseSuite) TestAwkwardOffsetBigChunk(c *C) {
	in, out := channels()
	go Transform(7, in, out)
	defer close(in)

	in <- []byte("orld hola mundo")
	in <- []byte("good bye!")

	res1 := <-out
	res2 := <-out
	c.Assert(string(res1), Equals, string(format(7, []byte("orld hola mundo"))))
	c.Assert(string(res2), Equals, string(format(22, []byte("good bye!"))))
}
