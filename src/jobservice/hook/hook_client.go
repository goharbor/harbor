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

package hook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/lib/log"
)

// Client for handling the hook events
type Client interface {
	// SendEvent send the event to the subscribed parties
	SendEvent(evt *Event) error
}

// Client is used to post the related data to the interested parties.
type basicClient struct {
	client *http.Client
	ctx    context.Context
}

// NewClient return the ptr of the new hook client
func NewClient(ctx context.Context) Client {
	// Create transport
	transport := &http.Transport{
		MaxIdleConns: 20,
		// hooks only ever target the single core endpoint, so allow all idle
		// connections to be kept for that host (default is 2, which causes
		// connection churn and TIME_WAIT pile-up under heavy job activity)
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     30 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Proxy:                 http.ProxyFromEnvironment,
	}
	if commonhttp.InternalEnableVerifyClientCert() {
		tlsConfig, err := commonhttp.GetInternalTLSConfig()
		if err != nil {
			log.Errorf("client load cert file with err: %v", err)
		}
		transport.TLSClientConfig = tlsConfig
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	return &basicClient{
		client: client,
		ctx:    ctx,
	}
}

// ReportStatus reports the status change info to the subscribed party.
// The status includes 'checkin' info with format 'check_in:<message>'
func (bc *basicClient) SendEvent(evt *Event) error {
	if evt == nil {
		return errors.New("nil event")
	}

	if err := evt.Validate(); err != nil {
		return err
	}

	// Marshal data
	data, err := json.Marshal(evt.Data)
	if err != nil {
		return err
	}

	// New post request
	req, err := http.NewRequest(http.MethodPost, rebaseOnCore(evt.URL), strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	res, err := bc.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}() // close connection for reuse

	// Should be 200
	if res.StatusCode != http.StatusOK {
		if res.ContentLength > 0 {
			// read error content and return
			dt, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}
			return errors.New(string(dt))
		}

		return fmt.Errorf("failed to report status change via hook, expect '200' but got '%d'", res.StatusCode)
	}

	return nil
}

// rebaseOnCore swaps the scheme and host of a hook URL for core's current
// address. The URL is frozen into the job record when the job is created, so it
// goes stale whenever CORE_URL changes under a running instance, e.g. when
// internalTLS is toggled and core moves from http:80 to https:443.
func rebaseOnCore(raw string) string {
	core, err := url.Parse(config.GetCoreURL())
	if err != nil || core.Host == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || !strings.HasPrefix(u.Path, "/service/notifications/") {
		return raw
	}
	u.Scheme, u.Host = core.Scheme, core.Host
	if rebased := u.String(); rebased != raw {
		log.Infof("hook URL %s rebased to %s: CORE_URL changed since the job was created", raw, rebased)
		return rebased
	}
	return raw
}
