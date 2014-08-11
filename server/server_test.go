package server

import (
	"bytes"
	"fmt"
	"github.com/naaman/busl/broker"
	. "github.com/naaman/busl/util"
	. "gopkg.in/check.v1"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type HttpServerSuite struct{}

var _ = Suite(&HttpServerSuite{})
var sf = fmt.Sprintf

func newRequest(method, url, body string) *http.Request {
	return newRequestFromReader(method, url, bytes.NewBufferString(body))
}

func newRequestFromReader(method, url string, reader io.Reader) *http.Request {
	request, _ := http.NewRequest(method, url, reader)
	urlParts := strings.Split(url, "/")
	if method == "POST" {
		request.TransferEncoding = []string{"chunked"}
		request.Header.Add("Transfer-Encoding", "chunked")
	}
	if len(urlParts) == 3 {
		streamId := urlParts[2]
		setStreamId(request, streamId)
	}
	return request
}

func setStreamId(req *http.Request, streamId string) {
	req.URL.RawQuery = "%3Auuid=" + streamId + "&"
}

func (s *HttpServerSuite) TestMkstream(c *C) {
	request := newRequest("POST", "/streams", "")
	response := httptest.NewRecorder()

	mkstream(response, request)

	c.Assert(response.Code, Equals, 200)
	c.Assert(response.Body.String(), HasLen, 32)
}

func (s *HttpServerSuite) TestPub(c *C) {
	request := newRequest("POST", "/streams/1234", "")
	response := httptest.NewRecorder()

	pub(response, request)

	c.Assert(response.Code, Equals, http.StatusOK)
	c.Assert(response.Body.String(), IsEmptyString)
}

func (s *HttpServerSuite) TestPubWithoutTransferEncoding(c *C) {
	request, _ := http.NewRequest("POST", "/streams/1234", nil)
	setStreamId(request, "1234")
	response := httptest.NewRecorder()

	pub(response, request)

	c.Assert(response.Code, Equals, http.StatusBadRequest)
	c.Assert(response.Body.String(), Equals, "A chunked Transfer-Encoding header is required.\n")
}

func (s *HttpServerSuite) TestSub(c *C) {
	streamId, _ := NewUUID()
	registrar := broker.NewRedisRegistrar()
	registrar.Register(streamId)
	publisher := broker.NewRedisBroker(streamId)

	request := newRequest("GET", sf("/streams/%s", streamId), "")
	response := CloseNotifierRecorder{httptest.NewRecorder(), make(chan bool, 1)}

	waiter := TimeoutFunc(time.Millisecond*5, func() {
		sub(response, request)
	})

	publisher.Publish([]byte("busl1"))
	publisher.UnsubscribeAll()
	<-waiter

	c.Assert(response.Code, Equals, http.StatusOK)
	c.Assert(response.Body.String(), Equals, "busl1")
}

func (s *HttpServerSuite) TestPubSub(c *C) {
	streamId, _ := NewUUID()
	registrar := broker.NewRedisRegistrar()
	registrar.Register(streamId)

	body := new(bytes.Buffer)
	bodyCloser := ioutil.NopCloser(body)

	pubRequest := newRequestFromReader("POST", sf("/streams/%s", streamId), bodyCloser)
	pubResponse := CloseNotifierRecorder{httptest.NewRecorder(), make(chan bool, 1)}

	pubBlocker := TimeoutFunc(time.Millisecond*5, func() {
		pub(pubResponse, pubRequest)
	})

	subRequest := newRequest("GET", sf("/streams/%s", streamId), "")
	subResponse := CloseNotifierRecorder{httptest.NewRecorder(), make(chan bool, 1)}

	subBlocker := TimeoutFunc(time.Millisecond*5, func() {
		sub(subResponse, subRequest)
	})

	for _, m := range []string{"first", " ", "second", " ", "third"} {
		body.Write([]byte(m))
	}

	bodyCloser.Close()
	<-pubBlocker
	<-subBlocker

	c.Assert(subResponse.Code, Equals, http.StatusOK)
	c.Assert(subResponse.Body.String(), Equals, "first second third")
}

type CloseNotifierRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (cnr CloseNotifierRecorder) close() {
	cnr.closed <- true
}

func (cnr CloseNotifierRecorder) CloseNotify() <-chan bool {
	return cnr.closed
}
