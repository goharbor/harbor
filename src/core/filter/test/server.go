package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/api/authentication/v1beta1"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// NewAuthProxyTestServer mocks a https server for auth proxy.
func NewAuthProxyTestServer() (*httptest.Server, error) {
	const webhookPath = "/authproxy/tokenreview"

	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unexpected method: %v", r.Method), http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != webhookPath {
			http.Error(w, fmt.Sprintf("unexpected path: %v", r.URL.Path), http.StatusNotFound)
			return
		}

		var review v1beta1.TokenReview
		bodyData, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(bodyData, &review); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode body: %v", err), http.StatusBadRequest)
			return
		}
		// ensure we received the serialized tokenreview as expected
		if review.APIVersion != "authentication.k8s.io/v1beta1" {
			http.Error(w, fmt.Sprintf("wrong api version: %s", string(bodyData)), http.StatusBadRequest)
			return
		}

		type userInfo struct {
			Username string              `json:"username"`
			UID      string              `json:"uid"`
			Groups   []string            `json:"groups"`
			Extra    map[string][]string `json:"extra"`
		}
		type status struct {
			Authenticated bool     `json:"authenticated"`
			User          userInfo `json:"user"`
			Audiences     []string `json:"audiences"`
		}

		var extra map[string][]string
		if review.Status.User.Extra != nil {
			extra = map[string][]string{}
			for k, v := range review.Status.User.Extra {
				extra[k] = v
			}
		}

		resp := struct {
			Kind       string `json:"kind"`
			APIVersion string `json:"apiVersion"`
			Status     status `json:"status"`
		}{
			Kind:       "TokenReview",
			APIVersion: v1beta1.SchemeGroupVersion.String(),
			Status: status{
				true,
				userInfo{
					Username: "administrator@vsphere.local",
					UID:      review.Status.User.UID,
					Groups:   review.Status.User.Groups,
					Extra:    extra,
				},
				review.Status.Audiences,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(serveHTTP))
	server.StartTLS()

	serverURL, _ := url.Parse(server.URL)
	serverURL.Path = webhookPath
	server.URL = serverURL.String()

	return server, nil
}
