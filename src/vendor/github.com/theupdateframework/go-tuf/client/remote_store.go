package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type HTTPRemoteOptions struct {
	MetadataPath string
	TargetsPath  string
	UserAgent    string
	Retries      *HTTPRemoteRetries
}

type HTTPRemoteRetries struct {
	Delay time.Duration
	Total time.Duration
}

var DefaultHTTPRetries = &HTTPRemoteRetries{
	Delay: time.Second,
	Total: 10 * time.Second,
}

func HTTPRemoteStore(baseURL string, opts *HTTPRemoteOptions, client *http.Client) (RemoteStore, error) {
	if !strings.HasPrefix(baseURL, "http") {
		return nil, ErrInvalidURL{baseURL}
	}
	if opts == nil {
		opts = &HTTPRemoteOptions{}
	}
	if opts.TargetsPath == "" {
		opts.TargetsPath = "targets"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &httpRemoteStore{baseURL, opts, client}, nil
}

type httpRemoteStore struct {
	baseURL string
	opts    *HTTPRemoteOptions
	cli     *http.Client
}

func (h *httpRemoteStore) GetMeta(name string) (io.ReadCloser, int64, error) {
	return h.get(path.Join(h.opts.MetadataPath, name))
}

func (h *httpRemoteStore) GetTarget(name string) (io.ReadCloser, int64, error) {
	return h.get(path.Join(h.opts.TargetsPath, name))
}

func (h *httpRemoteStore) get(s string) (io.ReadCloser, int64, error) {
	u := h.url(s)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, 0, err
	}
	if h.opts.UserAgent != "" {
		req.Header.Set("User-Agent", h.opts.UserAgent)
	}
	var res *http.Response
	if r := h.opts.Retries; r != nil {
		for start := time.Now(); time.Since(start) < r.Total; time.Sleep(r.Delay) {
			res, err = h.cli.Do(req)
			if err == nil && (res.StatusCode < 500 || res.StatusCode > 599) {
				break
			}
		}
	} else {
		res, err = h.cli.Do(req)
	}
	if err != nil {
		return nil, 0, err
	}

	if res.StatusCode == http.StatusNotFound {
		res.Body.Close()
		return nil, 0, ErrNotFound{s}
	} else if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, 0, &url.Error{
			Op:  "GET",
			URL: u,
			Err: fmt.Errorf("unexpected HTTP status %d", res.StatusCode),
		}
	}

	size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 0)
	if err != nil {
		return res.Body, -1, nil
	}
	return res.Body, size, nil
}

func (h *httpRemoteStore) url(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return h.baseURL + path
}
