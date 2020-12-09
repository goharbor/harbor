package exporter

import (
	"fmt"
	"net/http"
	"net/url"
)

var hbrCli *HarborClient

// HarborClient is client for request harbor
type HarborClient struct {
	HarborScheme string
	HarborHost   string
	HarborPort   int
	*http.Client
}

func (hc HarborClient) harborURL(p string) url.URL {
	return url.URL{
		Scheme: hc.HarborScheme,
		Host:   fmt.Sprintf("%s:%d", hc.HarborHost, hc.HarborPort),
		Path:   p,
	}
}

// Get ...
func (hc HarborClient) Get(p string) (*http.Response, error) {
	hbrURL := hc.harborURL(p)
	res, err := http.Get(hbrURL.String())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// InitHarborClient initialize the harbor client
func InitHarborClient(hc *HarborClient) {
	hbrCli = hc
}
