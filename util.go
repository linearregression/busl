package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
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

type StringSliceUtil []string

func (s StringSliceUtil) Contains(check string) bool {
	for _, c := range s {
		if c == check {
			return true
		}
	}
	return false
}

func count(metric string) { countMany(metric, 1) }
func countMany(metric string, count int64) { countWithData(metric, count, "") }

func countWithData(metric string, count int64, extraData string, v ...interface {}) {
	if extraData == "" {
		log.Printf("count#%s=%d", metric, count)
	} else {
		log.Printf("count#%s=%d %s", metric, count, extraData, v)
	}
}
