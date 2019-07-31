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

package swift

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/core/systeminfo/imagestorage"
	"github.com/goharbor/harbor/src/core/systeminfo/imagestorage/swift/internal"
	"github.com/gophercloud/gophercloud"
	"github.com/stretchr/testify/assert"
)

func newTestDriver(container string, handler http.HandlerFunc) (imagestorage.Driver, *httptest.Server, error) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case fmt.Sprintf("/%s", container):
			handler(w, r)
		case "/":
			internal.IdentityHandler(server, w, r)
		case "/v3/auth/tokens":
			internal.CatalogHandler(server.URL, w, r)
		}
	}))

	driver, err := NewDriver(gophercloud.AuthOptions{
		IdentityEndpoint: server.URL,
		TokenID:          "token",
	}, "region", container)

	return driver, server, err
}

func TestName(t *testing.T) {
	driver, server, err := newTestDriver("container", func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("X-Container-Meta-Quota-Bytes", "ABCDE")
		headers.Add("X-Container-Bytes-Used", "10000")
	})
	defer server.Close()
	assert.Nil(t, err, "unexpected error")

	assert.Equal(t, driver.Name(), DriverName, "unexpected total capacity")
}

func TestCap(t *testing.T) {
	driver, server, err := newTestDriver("container", func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("X-Container-Meta-Quota-Bytes", "15000")
		headers.Add("X-Container-Bytes-Used", "10000")
	})
	defer server.Close()
	assert.Nil(t, err, "unexpected error")

	cap, err := driver.Cap()
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, cap.Total, uint64(15000), "unexpected total capacity")
	assert.Equal(t, cap.Free, uint64(5000), "unexpected free capacity")
}

func TestCapNoQuota(t *testing.T) {
	driver, server, err := newTestDriver("container", func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("X-Container-Bytes-Used", "10000")
	})
	defer server.Close()
	assert.Nil(t, err, "unexpected error")

	cap, err := driver.Cap()
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, cap.Total, uint64(0), "unexpected total capacity")
	assert.Equal(t, cap.Free, uint64(0), "unexpected free capacity")
}

func TestCapNoBytesUsed(t *testing.T) {
	driver, server, err := newTestDriver("container", func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("X-Container-Meta-Quota-Bytes", "15000")
	})
	defer server.Close()
	assert.Nil(t, err, "unexpected error")

	cap, err := driver.Cap()
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, cap.Total, uint64(15000), "unexpected total capacity")
	assert.Equal(t, cap.Free, uint64(15000), "unexpected free capacity")
}

func TestCapInvalidBytesUsed(t *testing.T) {
	driver, server, err := newTestDriver("container", func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("X-Container-Meta-Quota-Bytes", "15000")
		headers.Add("X-Container-Bytes-Used", "AA")
	})
	defer server.Close()
	assert.Nil(t, err, "unexpected error")

	_, err = driver.Cap()
	assert.NotNil(t, err, "expecting error")
}

func TestOverflowed(t *testing.T) {
	driver, server, err := newTestDriver("container", func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("X-Container-Meta-Quota-Bytes", "10000")
		headers.Add("X-Container-Bytes-Used", "15000")
	})
	defer server.Close()
	assert.Nil(t, err, "unexpected error")

	cap, err := driver.Cap()
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, cap.Total, uint64(10000), "unexpected total capacity")
	assert.Equal(t, cap.Free, uint64(0), "unexpected free capacity")
}

func TestCapWithError(t *testing.T) {
	driver, server, err := newTestDriver("container", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()
	assert.Nil(t, err, "unexpected error")

	_, err = driver.Cap()
	assert.NotNil(t, err, "expecting error")
}
