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
)

func newPool(server *url.URL) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 4 * time.Minute,
		Dial: func () (redis.Conn, error) {
			c, err := redis.Dial("tcp", server.Host)
			if err != nil {
				return nil, err
			}
			if server.User != nil {
				if pw, _ := server.User.Password(); pw != "" {
					if _, err := c.Do("AUTH", pw); err != nil {
						c.Close()
						return nil, err
					}
				}
			}
			return c, err
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

func (b *RedisBroker) Subscribe() chan []byte {
	ch := make(chan []byte, msgBuf)
	b.subscribers[ch] = true
	go b.redisSubscribe(ch)
	return ch
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

	_, err := conn.Do("PUBLISH", channel, msg)
	if err != nil {
		log.Printf("publish: %s", err)
	}
}
