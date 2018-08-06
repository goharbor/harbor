package chartserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

//TestGetChartsByNs tests GetChartsByNs method in UtilityHandler
func TestGetChartsByNs(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/api/repo1/charts":
			if r.Method == http.MethodGet {
				w.Write(chartListContent)
				return
			}
		}

		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("not supported"))
	}))
	defer mockServer.Close()

	serverURL, err := url.Parse(mockServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	theController, err := NewController(serverURL)
	if err != nil {
		t.Fatal(err)
	}

	charts, err := theController.GetUtilityHandler().GetChartsByNs("repo1")
	if err != nil {
		t.Fatal(err)
	}

	if len(charts) != 2 {
		t.Fatalf("expect 2 items but got %d", len(charts))
	}
}
