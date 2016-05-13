package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/heroku/busl/broker"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestPublishAndSubscribe(t *testing.T) {
	handler := NewHandler(&Config{false, "", time.Second, ""})
	server := httptest.NewServer(handler)
	defer server.Close()

	data := [][]byte{
		[]byte{'h', 'e', 'l', 'l', 'o'},
		[]byte{0x1f, 0x8b, 0x08, 0x00, 0x3f, 0x6b, 0xe1, 0x53, 0x00, 0x03, 0xed, 0xce, 0xb1, 0x0a, 0xc2, 0x30},
		bytes.Repeat([]byte{'0'}, 32769),
	}

	for _, expected := range data {
		// uuid = curl -XPOST <url>/streams
		resp, err := http.Post(server.URL+"/streams", "", nil)
		defer resp.Body.Close()
		assert.Nil(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)

		// uuid extracted
		uuid := string(body)
		assert.Len(t, uuid, 36)

		done := make(chan bool)

		go func() {
			// curl <url>/streams/<uuid>
			// -- waiting for publish to arrive
			resp, err = http.Get(server.URL + "/streams/" + uuid)
			defer resp.Body.Close()
			assert.Nil(t, err)

			body, _ = ioutil.ReadAll(resp.Body)
			assert.Equal(t, body, expected)

			done <- true
		}()

		transport := &http.Transport{}
		client := &http.Client{Transport: transport}

		// curl -XPOST -H "Transfer-Encoding: chunked" -d "hello" <url>/streams/<uuid>
		req, _ := http.NewRequest("POST", server.URL+"/streams/"+uuid, bytes.NewReader(expected))
		req.TransferEncoding = []string{"chunked"}
		r, err := client.Do(req)
		r.Body.Close()
		assert.Nil(t, err)

		<-done

		// Read the whole response after the publisher has
		// completed. The mechanics of this is different in that
		// most of the content will be replayed instead of received
		// in chunks as they arrive.
		resp, err = http.Get(server.URL + "/streams/" + uuid)
		defer resp.Body.Close()
		assert.Nil(t, err)

		body, _ = ioutil.ReadAll(resp.Body)
		assert.Equal(t, body, expected)
	}
}

func TestPublishAndSubscribeWithSSE(t *testing.T) {
	handler := NewHandler(&Config{false, "", time.Second, ""})
	server := httptest.NewServer(handler)
	defer server.Close()

	data := []struct {
		offset int
		input  string
		output string
	}{
		{0, "hello", "id: 5\ndata: hello\n\n"},
		{0, "hello\n", "id: 6\ndata: hello\ndata: \n\n"},
		{0, "hello\nworld", "id: 11\ndata: hello\ndata: world\n\n"},
		{0, "hello\nworld\n", "id: 12\ndata: hello\ndata: world\ndata: \n\n"},
		{1, "hello\nworld\n", "id: 12\ndata: ello\ndata: world\ndata: \n\n"},
		{6, "hello\nworld\n", "id: 12\ndata: world\ndata: \n\n"},
		{11, "hello\nworld\n", "id: 12\ndata: \ndata: \n\n"},
		{12, "hello\nworld\n", ""},
	}

	client := &http.Client{Transport: &http.Transport{}}

	for _, testdata := range data {
		uuid := uuid.NewV4()
		url := fmt.Sprintf("%s/streams/%s", server.URL, uuid)

		// curl -XPUT <url>/streams/<uuid>
		request, _ := http.NewRequest("PUT", url, nil)
		resp, err := client.Do(request)
		defer resp.Body.Close()
		assert.Nil(t, err)

		done := make(chan bool)

		// curl -XPOST -H "Transfer-Encoding: chunked" -d "hello" <url>/streams/<uuid>
		req, _ := http.NewRequest("POST", url, bytes.NewReader([]byte(testdata.input)))
		req.TransferEncoding = []string{"chunked"}

		r, err := client.Do(req)
		assert.Nil(t, err)
		r.Body.Close()

		go func() {
			request, _ := http.NewRequest("GET", url, nil)
			request.Header.Add("Accept", "text/event-stream")
			request.Header.Add("Last-Event-Id", strconv.Itoa(testdata.offset))

			// curl <url>/streams/<uuid>
			// -- waiting for publish to arrive
			resp, err = client.Do(request)
			defer resp.Body.Close()
			assert.Nil(t, err)

			body, _ := ioutil.ReadAll(resp.Body)
			assert.Equal(t, fmt.Sprintf("%s", body), testdata.output)

			if len(body) == 0 {
				assert.Equal(t, resp.StatusCode, http.StatusNoContent)
			}

			done <- true
		}()

		<-done
	}
}

func TestCreateStream(t *testing.T) {
	handler := NewHandler(&Config{false, "", time.Second, ""})
	server := httptest.NewServer(handler)
	defer server.Close()

	transport := &http.Transport{}
	client := &http.Client{Transport: transport}

	// uuid = curl -XPUT <url>/streams/1/2/3
	request, _ := http.NewRequest("PUT", server.URL+"/streams/1/2/3", nil)
	resp, err := client.Do(request)
	defer resp.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusCreated)

	registrar := broker.NewRedisRegistrar()
	assert.True(t, registrar.IsRegistered("1/2/3"))
}

func TestSubGoneWithBackend(t *testing.T) {
	uuid := uuid.NewV4()

	storage, get, _ := fileServer(uuid.String())
	defer storage.Close()

	handler := NewHandler(&Config{false, "", time.Second, storage.URL})
	server := httptest.NewServer(handler)
	defer server.Close()

	get <- []byte("hello world")

	resp, err := http.Get(fmt.Sprintf("%s/streams/%s", server.URL, uuid))
	defer resp.Body.Close()
	assert.Nil(t, err)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, fmt.Sprintf("%s", body), "hello world")
}

func TestCreateStreamWithBackend(t *testing.T) {
	uuid := uuid.NewV4()

	storage, _, put := fileServer(uuid.String())
	defer storage.Close()

	handler := NewHandler(&Config{false, "", time.Second, storage.URL})
	server := httptest.NewServer(handler)
	defer server.Close()

	transport := &http.Transport{}
	client := &http.Client{Transport: transport}

	registrar := broker.NewRedisRegistrar()
	registrar.Register(uuid.String())

	// uuid = curl -XPUT <url>/streams/1/2/3
	request, _ := http.NewRequest("POST", fmt.Sprintf("%s/streams/%s", server.URL, uuid), bytes.NewReader([]byte("hello world")))
	request.TransferEncoding = []string{"chunked"}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, <-put, []byte("hello world"))
}

func TestAuthentication(t *testing.T) {
	handler := NewHandler(&Config{false, "u:pass1|u:pass2", time.Second, ""})
	server := httptest.NewServer(handler)
	defer server.Close()

	transport := &http.Transport{}
	client := &http.Client{Transport: transport}

	testdata := map[string]string{
		"POST": "/streams",
		"PUT":  "/streams/1/2/3",
	}

	status := map[string]int{
		"POST": http.StatusOK,
		"PUT":  http.StatusCreated,
	}

	// Validate that we return 401 for empty and invalid tokens
	for _, token := range []string{"", "invalid"} {
		for method, path := range testdata {
			request, _ := http.NewRequest(method, server.URL+path, nil)
			if token != "" {
				request.SetBasicAuth("", token)
			}
			resp, err := client.Do(request)
			defer resp.Body.Close()
			assert.Nil(t, err)
			assert.Equal(t, resp.Status, "401 Unauthorized",
				fmt.Sprintf("%s request to %s should be 401. Was %s", method, path, resp.Status))
		}
	}

	// Validate that all the colon separated token values are
	// accepted
	for _, token := range []string{"pass1", "pass2"} {
		for method, path := range testdata {
			request, _ := http.NewRequest(method, server.URL+path, nil)
			request.SetBasicAuth("u", token)
			resp, err := client.Do(request)
			defer resp.Body.Close()
			assert.Nil(t, err)
			assert.Equal(t, resp.StatusCode, status[method])
		}
	}
}

func fileServer(id string) (*httptest.Server, chan []byte, chan []byte) {
	get := make(chan []byte, 10)
	put := make(chan []byte, 10)

	mux := http.NewServeMux()
	mux.HandleFunc("/"+id, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write(<-get)
		case "PUT":
			b, _ := ioutil.ReadAll(r.Body)
			put <- b
		}
	})

	server := httptest.NewServer(mux)
	return server, get, put
}
