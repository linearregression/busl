package main

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
	"net/url"
	"os"
	"flag"
)

const (
	msgBuf = 10
)

var (
	redisUrl = flag.String("redisUrl", os.Getenv("REDIS_URL"), "URL of the redis server")
	redisServer, _ = url.Parse(*redisUrl)
	redisPool *redis.Pool = newPool(redisServer)
	redisKeyExpire = 60 // redis uses seconds for EXPIRE
)

func newPool(server *url.URL) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 4 * time.Minute,
		Dial: func () (c redis.Conn, err error) {
			c, err = redis.Dial("tcp", server.Host)
			if err != nil {
				return
			}

			if server.User == nil {
				return
			}

			pw, pwset := server.User.Password()
			if !pwset {
				return
			}

			if _, err = c.Do("AUTH", pw); err != nil {
				c.Close()
				return
			}
			return
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

type channel string

type RedisBroker struct {
	channelId UUID
	subscribers map[chan []byte]bool
	psc redis.PubSubConn
}

func NewRedisBroker(uuid UUID) *RedisBroker {
	broker := &RedisBroker{uuid, make(map[chan []byte]bool), redis.PubSubConn{}}

	return broker
}

func (b *RedisBroker) Subscribe() (ch chan []byte) {
	ch = make(chan []byte, msgBuf)
	b.subscribers[ch] = true
	go b.redisSubscribe(ch)
	return
}

func (b *RedisBroker) redisSubscribe(ch chan []byte) {
	conn := redisPool.Get()
	defer conn.Close()
	b.psc = redis.PubSubConn{conn}
	b.psc.PSubscribe(b.channelId + "*")
	for {
		switch msg := b.psc.Receive().(type) {
		case redis.PMessage:
			switch msg.Channel {
			case string(b.channelId) + "kill":
				log.Printf("RedisBroker.redisSubscribe.Channel.%s.kill=1", b.channelId)
				b.psc.PUnsubscribe(b.channelId + "*")
			case string(b.channelId):
				ch <- msg.Data
			}
		case redis.Subscription:
			log.Printf("RedisBroker.redisSubscribe.Channel.%s.%s=%d", msg.Channel, msg.Kind, msg.Count)
			if msg.Kind == "punsubscribe" || msg.Kind == "unsubscribe" {
				b.Unsubscribe(ch)
				return
			}
		case error:
			log.Printf("RedisBroker.redisSubscribe.RecieveError=1 err=%s", msg)
			return
		}
	}
}

func (b *RedisBroker) Unsubscribe(ch chan []byte) {
	delete(b.subscribers, ch)
	close(ch)
}

func (b *RedisBroker) UnsubscribeAll() {
	log.Printf("RedisBroker.UnsubscribeAll=%d", len(b.subscribers))
	b.publishOn([]byte{1}, string(b.channelId) + "kill")
}

func (b *RedisBroker) Publish(msg []byte) {
	b.publishOn(msg, string(b.channelId))
}

func (b *RedisBroker) publishOn(msg []byte, channel string) {
	conn := redisPool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	conn.Send("PUBLISH", channel, msg)
	conn.Send("APPEND", channel, msg)
	conn.Send("EXPIRE", channel, redisKeyExpire)
	_, err := conn.Do("EXEC")
	if err != nil {
		log.Printf("publish: %s", err)
	}
}
