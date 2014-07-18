package main

import (
	"errors"
	"flag"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/url"
	"os"
	"time"
)

const (
	msgBuf = 10
)

var (
	redisUrl                       = flag.String("redisUrl", os.Getenv("REDIS_URL"), "URL of the redis server")
	redisServer, _                 = url.Parse(*redisUrl)
	redisPool          *redis.Pool = newPool(redisServer)
	redisKeyExpire                 = 60 // redis uses seconds for EXPIRE
	redisChannelExpire             = redisKeyExpire * 5
)

func newPool(server *url.URL) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 4 * time.Minute,
		Dial: func() (c redis.Conn, err error) {
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
	channelId   UUID
	subscribers map[chan []byte]bool
	psc         redis.PubSubConn
}

func NewRedisBroker(uuid UUID) *RedisBroker {
	broker := &RedisBroker{uuid, make(map[chan []byte]bool), redis.PubSubConn{}}

	return broker
}

func (b *RedisBroker) Subscribe() (ch chan []byte, err error) {
	if !NewRedisRegistrar().IsRegistered(b.channelId) {
		return nil, errors.New("Channel is not registered.")
	}

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

	if err := b.replay(b.channelId, ch); err != nil {
		b.Unsubscribe(ch)
		return
	}

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
	b.publishOn([]byte{1}, string(b.channelId)+"kill")

	conn := redisPool.Get()
	defer conn.Close()

	conn.Do("SETEX", string(b.channelId)+"done", redisChannelExpire, []byte{1})
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
	conn.Send("DEL", channel+"done")

	_, err := conn.Do("EXEC")
	if err != nil {
		log.Printf("publish: %s", err)
	}
}

func (b *RedisBroker) replay(channel UUID, ch chan []byte) (err error) {
	conn := redisPool.Get()
	defer conn.Close()

	result, err := conn.Do("MGET", string(channel), string(channel)+"done")
	if err != nil {
		log.Printf("publish: %s", err)
		return
	}

	resultArray := result.([]interface{})
	buffer, channelDone := getRedisByteArray(resultArray[0]), getRedisByteArray(resultArray[1])

	if buffer != nil {
		ch <- buffer
	}

	if channelDone != nil && channelDone[0] == 1 {
		log.Printf("channelDone: %s", channelDone)
		return errors.New("Channel is done.")
	}

	return
}

func getRedisByteArray(v interface{}) []byte {
	if v != nil {
		return v.([]byte)
	}
	return nil
}

type RedisRegistrar struct{}

func NewRedisRegistrar() *RedisRegistrar {
	registrar := &RedisRegistrar{}

	return registrar
}

func (rr *RedisRegistrar) Register(channel UUID) (err error) {
	conn := redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("SETEX", channel, redisChannelExpire, make([]byte, 0))
	if err != nil {
		log.Printf("register: %s", err)
		return
	}
	return
}

func (rr *RedisRegistrar) IsRegistered(channel UUID) (registered bool) {
	conn := redisPool.Get()
	defer conn.Close()

	result, err := conn.Do("EXISTS", channel)
	if err != nil {
		log.Printf("register: %s", err)
		return false
	}

	return result.(int64) == 1
}
