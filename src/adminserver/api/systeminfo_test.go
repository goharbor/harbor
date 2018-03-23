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

package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/adminserver/systeminfo/imagestorage"
)

type fakeImageStorageDriver struct {
	capacity *imagestorage.Capacity
	err      error
}

func (f *fakeImageStorageDriver) Name() string {
	return "fake"
}

func (f *fakeImageStorageDriver) Cap() (*imagestorage.Capacity, error) {
	return f.capacity, f.err
}

func TestCapacity(t *testing.T) {
	cases := []struct {
		driver       imagestorage.Driver
		responseCode int
		capacity     *imagestorage.Capacity
	}{
		{&fakeImageStorageDriver{nil, errors.New("error")}, http.StatusInternalServerError, nil},
		{&fakeImageStorageDriver{&imagestorage.Capacity{100, 90}, nil}, http.StatusOK, &imagestorage.Capacity{100, 90}},
	}

	req, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	for _, c := range cases {
		imagestorage.GlobalDriver = c.driver
		w := httptest.NewRecorder()
		Capacity(w, req)
		assert.Equal(t, c.responseCode, w.Code, "unexpected response code")
		if c.responseCode == http.StatusOK {
			b, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Fatalf("failed to read from response body: %v", err)
			}
			capacity := &imagestorage.Capacity{}
			if err = json.Unmarshal(b, capacity); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			assert.Equal(t, c.capacity, capacity)
		}
	}
}
