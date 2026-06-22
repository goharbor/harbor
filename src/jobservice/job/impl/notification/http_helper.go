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

package notification

import (
	"context"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib"
)

const (
	secure   = "secure"
	insecure = "insecure"

	// Max retry has the same meaning as max fails.
	maxFails = "JOBSERVICE_WEBHOOK_JOB_MAX_RETRY"
	// http client timeout for webhook job(seconds).
	httpClientTimeout = "JOBSERVICE_WEBHOOK_JOB_HTTP_CLIENT_TIMEOUT"
)

var (
	// timeout records the timeout for http client
	timeout    time.Duration
	httpHelper *HTTPHelper
)

func init() {
	// default timeout is 3 seconds
	timeout = 3 * time.Second
	if envTimeout, exist := os.LookupEnv(httpClientTimeout); exist {
		t, err := strconv.ParseInt(envTimeout, 10, 64)
		if err != nil {
			logger.Warningf("Failed to parse timeout from environment, error: %v", err)
			return
		}

		timeout = time.Duration(t) * time.Second
		logger.Debugf("Set the http client timeout to %v for webhook job", timeout)
	}
}

// HTTPHelper in charge of sending notification messages to remote endpoint
type HTTPHelper struct {
	clients map[string]*http.Client
}

func init() {
	httpHelper = &HTTPHelper{
		clients: map[string]*http.Client{},
	}
	httpHelper.clients[secure] = &http.Client{
		Transport:     webhookTransport(false),
		Timeout:       timeout,
		CheckRedirect: noRedirect,
	}
	httpHelper.clients[insecure] = &http.Client{
		Transport:     webhookTransport(true),
		Timeout:       timeout,
		CheckRedirect: noRedirect,
	}
}

type contextKey string

const (
	useProxyKey contextKey = "useProxy"
)

var (
	dialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
)

func webhookTransport(insecure bool) http.RoundTripper {
	opts := []func(*http.Transport){
		func(tr *http.Transport) {
			tr.Proxy = http.ProxyFromEnvironment
			tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				if useProxy, ok := ctx.Value(useProxyKey).(bool); ok && useProxy {
					return dialer.DialContext(ctx, network, addr)
				}
				return lib.PublicDialContext(ctx, network, addr)
			}
		},
	}
	if insecure {
		opts = append(opts, commonhttp.WithInsecureSkipVerify(true))
	}
	return commonhttp.NewTransport(opts...)
}

func noRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}
