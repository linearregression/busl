package server

import (
	"io"
	"time"

	"github.com/heroku/busl/util"
)

type payload struct {
	p   []byte
	n   int
	err error
}

type keepAliveReader struct {
	r        io.Reader
	packet   []byte        // typically a null byte
	interval time.Duration // duration before sending an ack
	ch       chan *payload // where all the original reads go to
	done     <-chan bool   // closeNotifier
	eof      bool          // marked true when we hit EOF
}

func newKeepAliveReader(r io.Reader, packet []byte, interval time.Duration, done <-chan bool) io.ReadCloser {
	ch := make(chan *payload, 100)

	go func() {
		for {
			payload := &payload{p: make([]byte, 1024*32)}
			payload.n, payload.err = r.Read(payload.p)
			ch <- payload

			if payload.err != nil {
				break
			}
		}
	}()

	return &keepAliveReader{r: r, ch: ch, done: done, packet: packet, interval: interval}
}

func (r *keepAliveReader) Read(p []byte) (int, error) {
	if r.eof {
		return 0, io.EOF
	}

	timer := time.NewTimer(r.interval)
	defer timer.Stop()

	select {
	case payload := <-r.ch:
		if payload.n > 0 {
			copy(p, payload.p[0:payload.n])
		}

		if payload.err == io.EOF {
			r.eof = true
		}

		return payload.n, payload.err

	case <-timer.C:
		util.Count("server.sub.keepAlive")
		return copy(p, r.packet), nil

	case <-r.done:
		util.Count("server.sub.clientClosed")
		r.eof = true
		return 0, io.EOF
	}
}

func (r *keepAliveReader) Close() error {
	if closer, ok := r.r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
