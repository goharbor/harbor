// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
	"strings"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

func currPath() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Dir(f)
}

func serveFile(rw http.ResponseWriter, p string) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}

	_, err2 := rw.Write(data)
	if err2 != nil {
		http.Error(rw, err2.Error(), 500)
	}
}

type notificationHandler struct {
	id string
}

func (n *notificationHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	suffix := strings.TrimPrefix(req.URL.Path, "/v1/notifications/")
	if req.Method == http.MethodDelete {
		rw.WriteHeader(200)
	} else if req.Method == http.MethodGet {
		if strings.HasPrefix(suffix, n.id) {
			serveFile(rw, path.Join(currPath(), "notification.json"))
		} else {
			rw.WriteHeader(404)
		}
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type layerHandler struct {
	name string
}

func (l *layerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		data, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			http.Error(rw, err.Error(), 500)
		}
		layer := &models.ClairLayerEnvelope{}
		if err := json.Unmarshal(data, layer); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		rw.WriteHeader(http.StatusCreated)
	} else if req.Method == http.MethodGet {
		name := strings.TrimPrefix(req.URL.Path, "/v1/layers/")
		if name == l.name {
			serveFile(rw, path.Join(currPath(), "total-12.json"))
		} else {
			http.Error(rw, fmt.Sprintf("Invalid layer name: %s", name), http.StatusNotFound)
		}
	} else {
		http.Error(rw, "", http.StatusMethodNotAllowed)
	}
}

// NewMockServer ...
func NewMockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/namespaces", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			serveFile(rw, path.Join(currPath(), "ns.json"))
		} else {
			rw.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.Handle("/v1/notifications/", &notificationHandler{id: "ec45ec87-bfc8-4129-a1c3-d2b82622175a"})
	mux.Handle("/v1/layers", &layerHandler{name: "03adedf41d4e0ea1b2458546a5b4717bf5f24b23489b25589e20c692aaf84d19"})
	mux.Handle("/v1/layers/", &layerHandler{name: "03adedf41d4e0ea1b2458546a5b4717bf5f24b23489b25589e20c692aaf84d19"})
	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		log.Infof("method: %s, path: %s", req.Method, req.URL.Path)
		rw.WriteHeader(http.StatusNotFound)
	},
	)
	return httptest.NewServer(mux)
}
