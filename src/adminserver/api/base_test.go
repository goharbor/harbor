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

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleInternalServerError(t *testing.T) {
	w := httptest.NewRecorder()
	handleInternalServerError(w)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandleBadRequestError(t *testing.T) {
	w := httptest.NewRecorder()
	err := "error message"
	handleBadRequestError(w, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	handleUnauthorized(w)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusUnauthorized)
	}
}

func TestWriteJSONNilInterface(t *testing.T) {
	w := httptest.NewRecorder()

	if err := writeJSON(w, nil); err != nil {
		t.Errorf("Expected nil error, received: %v", err)
	}
}

func TestWriteJSONMarshallErr(t *testing.T) {
	// Tests capture json.Marshall error
	x := map[string]interface{}{
		"foo": make(chan int),
	}

	w := httptest.NewRecorder()

	if err := writeJSON(w, x); err == nil {
		t.Errorf("Expected %v error received: no no error", err)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	if err := writeJSON(w, "Pong"); err != nil {
		t.Errorf("Expected nil error, received: %v", err)
	}
}
