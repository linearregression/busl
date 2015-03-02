package broker

import (
	"io"
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/heroku/busl/util"
)

type writer struct {
	conn    redis.Conn
	channel channel
}

func NewWriter(uuid util.UUID) (io.WriteCloser, error) {
	if !NewRedisRegistrar().IsRegistered(uuid) {
		return nil, ErrNotRegistered
	}

	conn := redisPool.Get()
	return &writer{conn, channel(uuid)}, nil
}

func (w *writer) Close() error {
	w.conn.Send("MULTI")
	w.conn.Send("SETEX", w.channel.doneId(), redisChannelExpire, []byte{1})
	w.conn.Send("PUBLISH", w.channel.killId(), 1)
	w.conn.Do("EXEC")
	return w.conn.Close()
}

func (w *writer) Write(p []byte) (int, error) {
	w.conn.Send("MULTI")
	w.conn.Send("APPEND", w.channel.id(), p)
	w.conn.Send("EXPIRE", w.channel.id(), redisChannelExpire)
	w.conn.Send("DEL", w.channel.doneId())
	w.conn.Send("PUBLISH", w.channel.id(), 1)

	_, err := w.conn.Do("EXEC")
	return len(p), err
}

type reader struct {
	conn     redis.Conn
	channel  channel
	psc      redis.PubSubConn
	offset   int64
	replayed bool
	closed   bool
	mutex    *sync.Mutex
}

func NewReader(uuid util.UUID) (io.ReadCloser, error) {
	if !NewRedisRegistrar().IsRegistered(uuid) {
		return nil, ErrNotRegistered
	}

	psc := redis.PubSubConn{redisPool.Get()}
	channel := channel(uuid)
	psc.PSubscribe(channel.wildcardId())

	rd := &reader{
		conn:    redisPool.Get(),
		channel: channel,
		psc:     psc,
		mutex:   &sync.Mutex{}}

	return rd, nil
}

// TODO: decide what to do based on whence; whether we should
// return an error if whence != 0, or if we should try and
// respect the values.
func (r *reader) Seek(offset int64, whence int) (int64, error) {
	r.offset = offset
	return offset, nil
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.closed { // Don't read from a closed redigo connection
		return 0, io.EOF
	}

	if n, err := r.replay(p); n > 0 || err != nil {
		return n, err
	}

	switch msg := r.psc.Receive().(type) {
	case redis.Message:
	case redis.PMessage:
		return r.read(msg, p)
	case redis.Subscription:
	case error:
		util.CountWithData("RedisBroker.redisSubscribe.RecieveError", 1, "err=%s", msg)
		err = msg
		return
	}
	return
}

func (r *reader) replay(p []byte) (n int, err error) {
	var buf []byte

	if !r.replayed {
		r.replayed = true

		buf, err = r.fetch()
		n = len(buf)

		if n > 0 {
			r.offset += int64(n)
			copy(p, buf)
		}

		if err == io.EOF {
			util.Count("RedisBroker.replay.channelDone")
		}
	}

	return n, err
}

func (r *reader) read(msg redis.PMessage, p []byte) (n int, err error) {
	var buf []byte

	if msg.Channel == r.channel.id() {
		buf, err = r.fetch()

		if n = len(buf); n > 0 {
			copy(p, buf)
			r.offset += int64(n)
		}
	}

	if msg.Channel == r.channel.killId() || err == io.EOF {
		util.Count("RedisBroker.redisSubscribe.Channel.kill")
		r.Close()
		err = io.EOF
	}

	return n, err
}

func (r *reader) fetch() ([]byte, error) {
	r.conn.Send("MULTI")
	r.conn.Send("GETRANGE", r.channel.id(), r.offset, -1)
	r.conn.Send("EXISTS", r.channel.doneId())

	list, err := redis.Values(r.conn.Do("EXEC"))
	data, err := redis.Bytes(list[0], err)
	done, err := redis.Bool(list[1], err)

	if done {
		err = io.EOF
	}

	return data, err
}

func (r *reader) Close() error {
	if r.closed {
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.closed = true
	r.psc.Unsubscribe()
	r.psc.Close()

	return r.conn.Close()
}
