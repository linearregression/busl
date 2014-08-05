package util

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"fmt"
	"time"
)

type UUID string

func NewUUID() (UUID, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}

	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return UUID(hex.EncodeToString(uuid)), nil
}

type NullByte []byte

func GetNullByte() []byte {
	return new(NullByte).Get()
}

func (nb NullByte) Get() []byte {
	nb = []byte{0}
	return nb
}

type StringSliceUtil []string

func (s StringSliceUtil) Contains(check string) bool {
	for _, c := range s {
		if c == check {
			return true
		}
	}
	return false
}

func Count(metric string) { CountMany(metric, 1) }

func CountMany(metric string, count int64) { CountWithData(metric, count, "") }

func CountWithData(metric string, count int64, extraData string, v ...interface{}) {
	if extraData == "" {
		log.Printf("count#%s=%d", metric, count)
	} else {
		log.Printf("count#%s=%d %s", metric, count, fmt.Sprintf(extraData, v))
	}
}

func TimeoutFunc(d time.Duration, ƒ func()) (ch chan bool) {
	ch = make(chan bool)
	time.AfterFunc(d, func() {
		ch <- true
	})
	go func() {
		ƒ()
		ch <- true
	}()
	return ch
}
