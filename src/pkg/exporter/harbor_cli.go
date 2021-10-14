package exporter

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
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
		Host:   net.JoinHostPort(hc.HarborHost, strconv.Itoa(hc.HarborPort)),
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
