package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/heroku/busl/util"
	"github.com/stretchr/testify/assert"
)

var baseURL = *util.StorageBaseURL

func setup() (string, string) {
	put, err := url.Parse(os.Getenv("TEST_PUT_URL"))
	get, err := url.Parse(os.Getenv("TEST_GET_URL"))

	if err != nil {
		return "", ""
	}

	return put.RequestURI()[1:], get.RequestURI()[1:]
}

func TestURLResolution(t *testing.T) {
	*util.StorageBaseURL = "https://bucket.s3.amazonaws.com"
	defer func() {
		*util.StorageBaseURL = ""
	}()

	fmt.Println(absoluteURL("/1/2/3?foo=bar"))
	fmt.Println(absoluteURL("1/2/3?foo=bar"))

	//Output:
	// https://bucket.s3.amazonaws.com/1/2/3?foo=bar <nil>
	// https://bucket.s3.amazonaws.com/1/2/3?foo=bar <nil>
}

func Example_empty_storage_base_url() {
	*util.StorageBaseURL = ""
	defer func() {
		*util.StorageBaseURL = baseURL
	}()
	fmt.Println(absoluteURL("/1/2/3?foo=bar"))

	//Output:
	// <nil> No storage defined
}

func TestPutConnRefused(t *testing.T) {
	*util.StorageBaseURL = "http://localhost:0"
	defer func() {
		*util.StorageBaseURL = baseURL
	}()

	err := Put("1/2/3", nil)
	assert.Error(t, err)
}

func TestGetConnRefused(t *testing.T) {
	*util.StorageBaseURL = "http://localhost:0"
	defer func() {
		*util.StorageBaseURL = baseURL
	}()

	_, err := Get("1/2/3", 0)
	assert.Error(t, err)
}

func TestPutWithoutBaseURL(t *testing.T) {
	*util.StorageBaseURL = ""
	defer func() {
		*util.StorageBaseURL = baseURL
	}()

	err := Put("1/2/3", nil)
	assert.Equal(t, err, ErrNoStorage)
}

func TestGetWithoutBaseURL(t *testing.T) {
	*util.StorageBaseURL = ""
	defer func() {
		*util.StorageBaseURL = baseURL
	}()

	_, err := Get("1/2/3", 0)
	assert.Equal(t, err, ErrNoStorage)
}

func TestPut(t *testing.T) {
	requestURI, _ := setup()
	if requestURI == "" {
		t.Skip("No PUT URL supplied")
	}

	reader := strings.NewReader("hello")
	err := Put(requestURI, reader)
	assert.Error(t, err)
}

func TestGet(t *testing.T) {
	_, requestURI := setup()
	if requestURI == "" {
		t.Skip("No GET URL supplied")
	}

	data := []string{
		"hello",
		"ello",
		"llo",
		"lo",
		"o",
	}

	for offset, expected := range data {
		r, _ := Get(requestURI, int64(offset))
		if r != nil {
			defer r.(io.Closer).Close()
		}

		bytes, err := ioutil.ReadAll(r)
		assert.Nil(t, err)
		assert.Equal(t, expected, string(bytes))
	}
}

func TestGetIllegalOffset(t *testing.T) {
	_, requestURI := setup()
	if requestURI == "" {
		t.Skip("No GET URL supplied")
	}

	_, err := Get(requestURI, 5)

	if err == nil || err.Error() == "Expected 200, got 416" {
		t.Fatalf("%v != Expected 200, got 416", err)
	}
}
