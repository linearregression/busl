package storage

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/heroku/busl/util"
)

// Number of times we should retry a failed HTTP request.
const retries = 3

var (
	ErrNoStorage = errors.New("No storage defined")
	Err4xx       = errors.New("HTTP 4xx")
	Err5xx       = errors.New("HTTP 5xx")
	ErrRange     = errors.New("HTTP 416: Invalid Range")
)

// Stores the given reader onto the underlying blob storage
// with the given requestURI. The requestURI is resolved
// using the `STORAGE_BASE_URL` as the base.
//
// Retries transient errors `retries` number of times.
//
// Usage:
//
//   reader := strings.NewReader("hello")
//   requestURI := "1/2/3?X-Amz-Algorithm=...&..."
//   err := storage.Put(requestURI, reader)
//
func Put(requestURI string, reader io.Reader) (err error) {
	for i := retries; i > 0; i-- {
		err = put(requestURI, reader)

		// Break if we get any error other than Err5xx
		if err != Err5xx {
			return err
		}

		// Log the put retry
		util.Count("storage.Put.retry")
	}

	// We've ran out of retries
	util.Count("storage.Put.error")
	return err
}

func put(requestURI string, reader io.Reader) error {
	req, err := newRequest("PUT", requestURI, reader)
	if err != nil {
		return err
	}
	res, err := process(req)
	if res != nil {
		defer res.Body.Close()
	}
	return err
}

// Grabs the data stored in requestURI.
// The requestURI is resolved using the `STORAGE_BASE_URL` as the base.
//
// Retries transient errors `retries` number of times.
//
// Usage:
//
//   requestURI := "1/2/3?X-Amz-Algorithm=...&..."
//   reader, err := storage.Get(requestURI, 0)
//
func Get(requestURI string, offset int64) (rd io.ReadCloser, err error) {
	for i := retries; i > 0; i-- {
		rd, err = get(requestURI, offset)

		// Break if we get:
		//   1) No errors
		//   2) Any error other than Err5xx
		if err == nil || err != Err5xx {
			return rd, err
		}

		// Close the body immediately to prevent
		// file descriptor leaks.
		if rd != nil {
			rd.Close()
		}

		util.Count("storage.Get.retry")
	}

	// We've ran out of retries
	util.Count("storage.Get.error")
	return rd, err
}

func get(requestURI string, offset int64) (io.ReadCloser, error) {
	req, err := newRequest("GET", requestURI, nil)
	if err != nil {
		return nil, err
	}
	req.TransferEncoding = []string{"chunked"}
	req.Header.Add("Transfer-Encoding", "chunked")
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-", offset))

	res, err := process(req)
	if res == nil {
		return nil, err
	}
	return res.Body, err
}

// constructs an http.Request object, resolving requestURI
// under `STORAGE_BASE_URL`.
func newRequest(method, requestURI string, reader io.Reader) (*http.Request, error) {
	u, err := absoluteURL(requestURI)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(method, u.String(), reader)
}

// Executes the HTTP request:
// Errors:
//
//   - Err5xx
//   - Err4xx
//   - ErrRange
//
func process(req *http.Request) (*http.Response, error) {
	transport := &http.Transport{}
	client := &http.Client{Transport: transport}

	res, err := client.Do(req)
	if err == nil {
		switch {
		case res.StatusCode == 416:
			err = ErrRange
		case res.StatusCode >= 500:
			err = Err5xx
		case res.StatusCode >= 400:
			err = Err4xx
		}
	}
	return res, err
}

func absoluteURL(requestURI string) (*url.URL, error) {
	if ref, err := url.ParseRequestURI(normalize(requestURI)); err != nil {
		return nil, err
	} else if uri, err := baseURI(); err != nil {
		return nil, err
	} else {
		return uri.ResolveReference(ref), nil
	}
}

func baseURI() (*url.URL, error) {
	if *util.StorageBaseUrl == "" {
		return nil, ErrNoStorage
	}
	return url.Parse(*util.StorageBaseUrl)
}

// Keep concept of requestURI similar to S3 without a slash prefix.
// We add the prefix to make `ParseRequestURI` happy.
func normalize(requestURI string) string {
	return "/" + requestURI
}
