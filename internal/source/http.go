package source

import (
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/lib/etag"
	"io"
	"net/http"
	"net/url"
)

const HttpScheme = "http"

const HttpsScheme = "https"

const httpClient = "http"

type Http struct {
	Client *http.Client
}

var _ Client = &Http{}

var ErrHttpClient = &Error{msg: "http client error"}

func NewHttpClient() *Http {
	return &Http{
		Client: http.DefaultClient,
	}
}

// Open returns a Client for the provided url.
// When Open is invoked, the client will send a HEAD request. Any error except HTTP Method Not Found from the HEAD request will be passed to the caller.
// The returned Client will send a GET request when Read is called for the first time
func (c *Http) Open(u *url.URL) (Source, error) {
	urlStr := u.String()

	// HEAD request
	resp, err := c.Client.Head(urlStr)
	if err != nil {
		return nil, ErrHttpClient.WithCause(err)
	}

	// Close body immediately because there is no content to expect
	err = resp.Body.Close()
	if err != nil {
		return nil, ErrHttpClient.WithCause(err)
	}

	if resp.StatusCode == http.StatusMethodNotAllowed {
		// When status is "405 Method Not Allowed", perform a GET without reading the response

		// GET request
		resp, err = c.Client.Get(urlStr)
		if err != nil {
			return nil, ErrHttpClient.WithCause(err)
		}

		// Keep the response body open because the returned source read from it
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrHttpClient.WithCause(fmt.Errorf("status %d %s", resp.StatusCode, resp.Status))
	}

	s := &httpSource{
		url: urlStr,
		reqFunc: func() (*http.Response, error) {
			return c.Client.Get(urlStr)
		},
		resp: resp,
	}

	return s, nil
}

func (c *Http) Schemes() []string {
	return []string{
		HttpScheme,
		HttpsScheme,
	}
}

type httpSource struct {
	url string

	// reqFunc performs the GET request
	reqFunc func() (*http.Response, error)

	// resp holds the *http.Response from the GET request
	resp *http.Response

	closed bool
}

var _ Source = &httpSource{}

func (s *httpSource) Read(p []byte) (int, error) {
	if s.closed {
		return 0, io.EOF
	}

	if s.resp == nil || s.resp.Request.Method == http.MethodHead {
		// Perform lazy GET request
		resp, err := s.reqFunc()
		if err != nil {
			return 0, err
		}

		// Replace response
		s.resp = resp
	}

	return s.resp.Body.Read(p)
}

func (s *httpSource) Close() error {
	s.closed = true

	if s.resp != nil {
		err := s.resp.Body.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *httpSource) Meta() (Meta, error) {
	if s.resp == nil {
		return nil, ErrHttpClient.WithCause(fmt.Errorf("http response unavailable"))
	}

	m := &httpMeta{
		meta{
			client: httpClient,
			url:    s.url,
			size:   s.resp.ContentLength,
			etag:   "",
		},
	}

	etagHeader := s.resp.Header.Get(etag.HeaderKey)
	if etagHeader != "" {
		etagStruct, err := etag.Parse(etagHeader)
		if err == nil && !etagStruct.Weak {
			// Ignore invalid or weak ETags
			m.etag = etagStruct.ETag
		}
	}

	return m, nil
}

type httpMeta struct {
	meta
}

var _ Meta = &httpMeta{}
