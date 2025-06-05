// Copyright Project Harbor Authors
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

package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/notification"
)

// This is a package to provide mock utilities.
var preheatMap = make(map[string]struct{})

// MockDragonflyProvider mocks a Dragonfly server.
func MockDragonflyProvider() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case dragonflyHealthPath:
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			w.WriteHeader(http.StatusOK)
		case dragonflyJobPath:
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			var resp = &dragonflyJobResponse{
				ID:        0,
				State:     dragonflyJobPendingState,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			bytes, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
		case fmt.Sprintf("%s/%s", dragonflyJobPath, "0"):
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			var resp = &dragonflyJobResponse{
				ID:        1,
				State:     dragonflyJobSuccessState,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			bytes, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		case fmt.Sprintf("%s/%s", dragonflyJobPath, "1"):
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			var resp = &dragonflyJobResponse{
				ID:        1,
				State:     dragonflyJobPendingState,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			bytes, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
		case fmt.Sprintf("%s/%s", dragonflyJobPath, "2"):
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			var resp = &dragonflyJobResponse{
				ID:        2,
				State:     dragonflyJobSuccessState,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			bytes, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
		case fmt.Sprintf("%s/%s", dragonflyJobPath, "3"):
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			var resp = &dragonflyJobResponse{
				ID:        3,
				State:     dragonflyJobFailureState,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			bytes, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if _, err := w.Write(bytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}))
}

// MockKrakenProvider mocks a Kraken server.
func MockKrakenProvider() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case krakenHealthPath:
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			w.WriteHeader(http.StatusOK)
		case krakenPreheatPath:
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			data, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			var payload = &notification.Notification{
				Events: []notification.Event{},
			}

			if err := json.Unmarshal(data, payload); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if len(payload.Events) > 0 {
				w.WriteHeader(http.StatusOK)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
}
