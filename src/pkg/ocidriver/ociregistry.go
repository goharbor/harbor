package ocidriver

import (
	"encoding/json"
	"fmt"
	http2 "github.com/goharbor/harbor/src/common/http"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// OciRegistry ...
type OciRegistry struct {
	baseURL   string
	transport http.RoundTripper
}

// NewRegistry ...
func NewRegistry(baseURL string, transport http.RoundTripper) (*OciRegistry, error) {
	return &OciRegistry{
		baseURL:   baseURL,
		transport: transport,
	}, nil
}

// Catalog ...
func (r *OciRegistry) Catalog( /*filter*/ ) ([]*OciRepository, error) {
	repos := []string{}
	aurl := buildCatalogURL(r.baseURL, 1000)

	for len(aurl) > 0 {
		resp, b, err := r.send("GET", aurl, nil)
		if err != nil {
			return nil, err
		}

		catalogResp := struct {
			Repositories []string `json:"repositories"`
		}{}

		if err := json.Unmarshal(b, &catalogResp); err != nil {
			return nil, err
		}

		repos = append(repos, catalogResp.Repositories...)
		// Link: </v2/_catalog?last=library%2Fhello-world-25&n=100>; rel="next"
		// Link: <http://domain.com/v2/_catalog?last=library%2Fhello-world-25&n=100>; rel="next"
		link := resp.Header.Get("Link")
		if strings.HasSuffix(link, `rel="next"`) && strings.Index(link, "<") >= 0 && strings.Index(link, ">") >= 0 {
			aurl = link[strings.Index(link, "<")+1 : strings.Index(link, ">")]
			if strings.Index(aurl, ":") < 0 {
				aurl = r.baseURL + aurl
			}
		} else {
			aurl = ""
		}

	}
	var rs []*OciRepository
	for _, repo := range repos {
		rs = append(rs, &OciRepository{
			Registry: r,
			Name:     repo,
		})
	}
	return rs, nil
}

// GetRepository ...
func (r *OciRegistry) GetRepository(name string) (*OciRepository, error) {
	return &OciRepository{
		Registry: r,
		Name:     name,
	}, nil
}

func buildCatalogURL(endpoint string, size int) string {
	return fmt.Sprintf("%s/v2/_catalog?n=%d", endpoint, size)
}

func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}

// Ping ...
func (r *OciRegistry) Ping() error {
	_, _, err := r.send("GET", buildPingURL(r.baseURL), nil)
	return err
}

func parseError(err error) error {
	if urlErr, ok := err.(*url.Error); ok {
		if regErr, ok := urlErr.Err.(*http2.Error); ok {
			return regErr
		}
	}
	return err
}

func (r *OciRegistry) send(method, url string, body io.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}
	client := http.Client{Transport: r.transport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, parseError(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return resp, b, nil
	}

	return nil, nil, &http2.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}
