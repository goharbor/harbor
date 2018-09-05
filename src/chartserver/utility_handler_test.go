package chartserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestGetChartsByNs tests GetChartsByNs method in UtilityHandler
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

// Test the function DeleteChart
func TestDeleteChart(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/api/repo1/charts/harbor":
			if r.Method == http.MethodGet {
				w.Write([]byte(chartVersionsOfHarbor))
				return
			}
		case "/api/repo1/charts/harbor/0.2.0",
			"/api/repo1/charts/harbor/0.2.1":
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusOK)
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

	if err := theController.GetUtilityHandler().DeleteChart("repo1", "harbor"); err != nil {
		t.Fatal(err)
	}
}

var chartVersionsOfHarbor = `
[
    {
        "name": "harbor",
        "home": "https://github.com/vmware/harbor",
        "sources": [
            "https://github.com/vmware/harbor/tree/master/contrib/helm/harbor"
        ],
        "version": "0.2.1",
        "description": "An Enterprise-class Docker Registry by VMware",
        "keywords": [
            "vmware",
            "docker",
            "registry",
            "harbor"
        ],
        "maintainers": [
            {
                "name": "Jesse Hu",
                "email": "huh@vmware.com"
            },
            {
                "name": "paulczar",
                "email": "username.taken@gmail.com"
            }
        ],
        "engine": "gotpl",
        "icon": "https://raw.githubusercontent.com/vmware/harbor/master/docs/img/harbor_logo.png",
        "appVersion": "1.5.0",
        "urls": [
            "charts/harbor-0.2.1.tgz"
        ],
        "created": "2018-08-29T10:26:29.625749155Z",
        "digest": "2538edf4ddb797af8e025f3bd6226270440110bbdb689bad48656a519a154236"
    },
    {
        "name": "harbor",
        "home": "https://github.com/vmware/harbor",
        "sources": [
            "https://github.com/vmware/harbor/tree/master/contrib/helm/harbor"
        ],
        "version": "0.2.0",
        "description": "An Enterprise-class Docker Registry by VMware",
        "keywords": [
            "vmware",
            "docker",
            "registry",
            "harbor"
        ],
        "maintainers": [
            {
                "name": "Jesse Hu",
                "email": "huh@vmware.com"
            },
            {
                "name": "paulczar",
                "email": "username.taken@gmail.com"
            }
        ],
        "engine": "gotpl",
        "icon": "https://raw.githubusercontent.com/vmware/harbor/master/docs/img/harbor_logo.png",
        "appVersion": "1.5.0",
        "urls": [
            "charts/harbor-0.2.0.tgz"
        ],
        "created": "2018-08-29T10:26:21.141611102Z",
        "digest": "fc8aae8dade9f0dfca12e9f1085081c49843d30a063a3fa7eb42497e3ceb277c"
    }
]
`
