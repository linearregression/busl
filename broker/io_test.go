package broker

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/heroku/busl/util"
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

func ExampleSeekCorrect() {
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
	fmt.Printf("%s", buf)

	//Output:
	// world
}

func ExampleSeekBeyond() {
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
	fmt.Printf("%s", buf)

	//Output:
	//
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
