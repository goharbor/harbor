package chartserver

import (
	"net/http"
	"testing"
)

// Test the URL rewrite function
func TestURLRewrite(t *testing.T) {
	req, err := createRequest(http.MethodGet, "/api/chartrepo/health")
	if err != nil {
		t.Fatal(err)
	}
	rewriteURLPath(req)
	if req.URL.Path != "/health" {
		t.Fatalf("Expect url format %s but got %s", "/health", req.URL.Path)
	}

	req, err = createRequest(http.MethodGet, "/api/chartrepo/library/charts")
	if err != nil {
		t.Fatal(err)
	}
	rewriteURLPath(req)
	if req.URL.Path != "/api/library/charts" {
		t.Fatalf("Expect url format %s but got %s", "/api/library/charts", req.URL.Path)
	}

	req, err = createRequest(http.MethodPost, "/api/chartrepo/charts")
	if err != nil {
		t.Fatal(err)
	}
	rewriteURLPath(req)
	if req.URL.Path != "/api/library/charts" {
		t.Fatalf("Expect url format %s but got %s", "/api/library/charts", req.URL.Path)
	}

	req, err = createRequest(http.MethodGet, "/chartrepo/library/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	rewriteURLPath(req)
	if req.URL.Path != "/library/index.yaml" {
		t.Fatalf("Expect url format %s but got %s", "/library/index.yaml", req.URL.Path)
	}
}

func createRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.RequestURI = url

	return req, nil
}
