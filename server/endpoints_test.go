package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestMkstream(t *testing.T) {
	endpoint := &Endpoint{&Config{false, "", time.Second, ""}}
	request, _ := http.NewRequest("POST", "/streams", nil)
	response := httptest.NewRecorder()

	endpoint.MakeUUID(response, request)

	assert.Equal(t, response.Code, 200)
	assert.Len(t, response.Body.String(), 36)
}

func TestSubscriber410(t *testing.T) {
	endpoint := &Endpoint{&Config{false, "", time.Second, ""}}
	streamID := uuid.NewV4()
	request, _ := http.NewRequest("GET", fmt.Sprintf("/streams/%s", streamID), nil)
	response := httptest.NewRecorder()

	endpoint.Subscriber(response, request)

	assert.Equal(t, response.Code, http.StatusNotFound)
	assert.Equal(t, response.Body.String(), "Channel is not registered.\n")
}

func TestPubNotRegistered(t *testing.T) {
	endpoint := &Endpoint{&Config{false, "", time.Second, ""}}
	streamID := uuid.NewV4()
	request, _ := http.NewRequest("POST", fmt.Sprintf("/streams/%s", streamID), nil)
	request.TransferEncoding = []string{"chunked"}
	response := httptest.NewRecorder()

	endpoint.Publisher(response, request)

	assert.Equal(t, response.Code, http.StatusNotFound)
}

func TestPubWithoutTransferEncoding(t *testing.T) {
	endpoint := &Endpoint{&Config{false, "", time.Second, ""}}
	request, _ := http.NewRequest("POST", "/streams/1234", nil)
	response := httptest.NewRecorder()

	endpoint.Publisher(response, request)

	assert.Equal(t, response.Code, http.StatusBadRequest)
	assert.Equal(t, response.Body.String(), "A chunked Transfer-Encoding header is required.\n")
}
