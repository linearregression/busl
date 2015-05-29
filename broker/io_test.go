package broker

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/heroku/busl/util"
	"github.com/stretchr/testify/assert"
)

func setup() string {
	registrar := NewRedisRegistrar()
	uuid, _ := util.NewUUID()
	registrar.Register(uuid)

	return uuid
}

func ExamplePubSub() {
	uuid := setup()

	r, _ := NewReader(uuid)
	defer r.(io.Closer).Close()

	pub := make(chan bool)
	done := make(chan bool)

	go func() {
		<-pub
		io.Copy(os.Stdout, r)
		done <- true
	}()

	go func() {
		pub <- true

		w, _ := NewWriter(uuid)
		w.Write([]byte("busl"))
		w.Write([]byte(" hello"))
		w.Write([]byte(" world"))
		w.Close()
	}()

	<-done

	//Output:
	// busl hello world
}

func ExampleFullReplay() {
	uuid := setup()

	w, _ := NewWriter(uuid)
	w.Write([]byte("busl"))
	w.Write([]byte(" hello"))
	w.Write([]byte(" world"))

	r, _ := NewReader(uuid)
	defer r.(io.Closer).Close()

	buf := make([]byte, 16)
	io.ReadAtLeast(r, buf, 16)

	fmt.Printf("%s", buf)

	//Output:
	// busl hello world
}

func TestSeekCorrect(t *testing.T) {
	uuid := setup()

	w, _ := NewWriter(uuid)
	w.Write([]byte("busl"))
	w.Write([]byte(" hello"))
	w.Write([]byte(" world"))
	w.Close()

	r, _ := NewReader(uuid)
	r.(io.Seeker).Seek(10, 0)
	defer r.(io.Closer).Close()

	buf, _ := ioutil.ReadAll(r)
	assert.Equal(t, " world", string(buf))
}

func TestSeekBeyond(t *testing.T) {
	uuid := setup()

	w, _ := NewWriter(uuid)
	w.Write([]byte("busl"))
	w.Write([]byte(" hello"))
	w.Write([]byte(" world"))
	w.Close()

	r, _ := NewReader(uuid)
	r.(io.Seeker).Seek(16, 0)
	defer r.Close()

	buf, _ := ioutil.ReadAll(r)
	assert.Equal(t, []byte{}, buf)
}

func ExampleHalfReplayHalfSubscribed() {
	uuid := setup()

	w, _ := NewWriter(uuid)
	w.Write([]byte("busl"))

	r, _ := NewReader(uuid)

	pub := make(chan bool)
	done := make(chan bool)

	go func() {
		<-pub
		io.Copy(os.Stdout, r)
		done <- true
	}()

	go func() {
		pub <- true

		w.Write([]byte(" hello"))
		w.Write([]byte(" world"))
		w.Close()
	}()

	<-done

	//Output:
	// busl hello world
}

func TestOverflowingBuffer(t *testing.T) {
	uuid := setup()

	w, _ := NewWriter(uuid)
	w.Write(bytes.Repeat([]byte("0"), 4096))
	w.Write(bytes.Repeat([]byte("1"), 4096))
	w.Write(bytes.Repeat([]byte("2"), 4096))
	w.Write(bytes.Repeat([]byte("3"), 4096))
	w.Write(bytes.Repeat([]byte("4"), 4096))
	w.Write(bytes.Repeat([]byte("5"), 4096))
	w.Write(bytes.Repeat([]byte("6"), 4096))
	w.Write(bytes.Repeat([]byte("7"), 4096))
	w.Write(bytes.Repeat([]byte("A"), 1))

	r, _ := NewReader(uuid)
	defer r.(io.Closer).Close()

	done := make(chan int64)
	go func() {
		n, _ := io.Copy(ioutil.Discard, r)
		done <- n
	}()
	w.Close()
	assert.Equal(t, int64(32769), <-done)
}
