package util

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"os/signal"
)

// NewUUID creates a random UUID
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}

	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid), nil
}

// StringInSlice checks a string array contains a check value
func StringInSlice(content []string, check string) bool {
	for _, c := range content {
		if c == check {
			return true
		}
	}
	return false
}

// AwaitSignals sets up a channel to wait for an unix signal
func AwaitSignals(signals ...os.Signal) <-chan struct{} {
	s := make(chan os.Signal, 1)
	signal.Notify(s, signals...)
	log.Printf("signals.await signals=%v\n", signals)

	received := make(chan struct{})
	go func() {
		log.Printf("signals.received signal=%v\n", <-s)
		close(received)
	}()

	return received
}
