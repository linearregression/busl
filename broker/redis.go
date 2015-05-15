package broker

import (
	"flag"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/heroku/busl/Godeps/_workspace/src/github.com/garyburd/redigo/redis"
	"github.com/heroku/busl/util"
)

var (
	redisUrl           = flag.String("redisUrl", os.Getenv("REDIS_URL"), "URL of the redis server")
	redisServer        *url.URL
	redisPool          *redis.Pool
	redisKeyExpire     = 60 // redis uses seconds for EXPIRE
	redisChannelExpire = redisKeyExpire * 5
)

func init() {
	flag.Parse()
	redisServer, _ = url.Parse(*redisUrl)
	redisPool = newPool(redisServer)

	conn := redisPool.Get()
	defer conn.Close()
}

func newPool(server *url.URL) *redis.Pool {
	cleanServerURL := *server
	cleanServerURL.User = nil
	log.Printf("connecting to redis: %s", cleanServerURL)
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

func (c channel) id() string {
	return string(c) + ":id"
}

func (c channel) wildcardId() string {
	return string(c) + ":*"
}

func (c channel) doneId() string {
	return string(c) + ":done"
}

func (c channel) killId() string {
	return string(c) + ":kill"
}

type RedisRegistrar struct{}

func NewRedisRegistrar() *RedisRegistrar {
	registrar := &RedisRegistrar{}

	return registrar
}

func (rr *RedisRegistrar) Register(channelName string) (err error) {
	conn := redisPool.Get()
	defer conn.Close()

	channel := channel(channelName)

	_, err = conn.Do("SETEX", channel.id(), redisChannelExpire, make([]byte, 0))
	if err != nil {
		util.CountWithData("RedisRegistrar.Register.error", 1, "error=%s", err)
		return
	}
	return
}

func (rr *RedisRegistrar) IsRegistered(channelName string) (registered bool) {
	conn := redisPool.Get()
	defer conn.Close()

	channel := channel(channelName)

	exists, err := redis.Bool(conn.Do("EXISTS", channel.id()))
	if err != nil {
		util.CountWithData("RedisRegistrar.IsRegistered.error", 1, "error=%s", err)
		return false
	}

	return exists
}

func Get(key string) ([]byte, error) {
	conn := redisPool.Get()
	defer conn.Close()

	channel := channel(key)
	return redis.Bytes(conn.Do("GET", channel.id()))
}
