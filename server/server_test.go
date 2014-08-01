package server

import (
	"bytes"
	"fmt"
	"github.com/naaman/busl/broker"
	. "github.com/naaman/busl/util"
	. "gopkg.in/check.v1"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type HttpServerSuite struct{}

var _ = Suite(&HttpServerSuite{})

func (s *HttpServerSuite) TestMkstream(c *C) {
	request, _ := http.NewRequest("POST", "/streams", nil)
	response := httptest.NewRecorder()

	mkstream(response, request)

	c.Assert(response.Code, Equals, 200)
	c.Assert(response.Body.String(), HasLen, 32)
}

func (s *HttpServerSuite) TestPub(c *C) {
	request, _ := http.NewRequest("POST", "/streams/1234", bytes.NewBufferString("busl2"))
	request.URL.RawQuery = "%3Auuid=1234&"
	request.TransferEncoding = []string{"chunked"}
	request.Header.Add("Transfer-Encoding", "chunked")
	response := httptest.NewRecorder()

	pub(response, request)

	c.Assert(response.Code, Equals, http.StatusOK)
	c.Assert(response.Body.String(), IsEmptyString)
}

func (s *HttpServerSuite) TestPubWithoutTransferEncoding(c *C) {
	request, _ := http.NewRequest("POST", "/streams/1234", nil)
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
	publisher.Publish([]byte("busl1\n"))

	request, _ := http.NewRequest("GET", fmt.Sprintf("/streams/%s", streamId), nil)
	request.URL.RawQuery = "%3Auuid=" + string(streamId) + "&"
	response := CloseNotifierRecorder{httptest.NewRecorder(), make(chan bool, 1)}

	time.AfterFunc(time.Millisecond*50, func() {
		response.close()
	})
	sub(response, request)

	c.Assert(response.Code, Equals, http.StatusOK)
	c.Assert(response.Body.String(), Equals, "busl1\n")
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
