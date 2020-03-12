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

package v1

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/scan/rest/auth"
	"github.com/pkg/errors"
)

const (
	// defaultRefreshInterval is the default interval with seconds of refreshing report
	defaultRefreshInterval = 5
	// refreshAfterHeader provides the refresh interval value
	refreshAfterHeader = "Refresh-After"
)

// Client defines the methods to access the adapter services that
// implement the REST API specs
type Client interface {
	// GetMetadata gets the metadata of the given scanner
	//
	//   Returns:
	//     *ScannerAdapterMetadata : metadata of the given scanner
	//     error                   : non nil error if any errors occurred
	GetMetadata() (*ScannerAdapterMetadata, error)

	// SubmitScan initiates a scanning of the given artifact.
	// Returns `nil` if the request was accepted, a non `nil` error otherwise.
	//
	//   Arguments:
	//     req *ScanRequest : request including the registry and artifact data
	//
	//   Returns:
	//     *ScanResponse : response with UUID for tracking the scan results
	//     error         : non nil error if any errors occurred
	SubmitScan(req *ScanRequest) (*ScanResponse, error)

	// GetScanReport gets the scan result for the corresponding ScanRequest identifier.
	// Note that this is a blocking method which either returns a non `nil` scan report or error.
	// A caller is supposed to cast the returned interface{} to a structure that corresponds
	// to the specified MIME type.
	//
	//   Arguments:
	//     scanRequestID string  : the ID of the scan submitted before
	//     reportMIMEType string : the report mime type
	//   Returns:
	//     string : the scan report of the given artifact
	//     error  : non nil error if any errors occurred
	GetScanReport(scanRequestID, reportMIMEType string) (string, error)
}

// basicClient is default implementation of the Client interface
type basicClient struct {
	httpClient *http.Client
	spec       *Spec
	authorizer auth.Authorizer
}

// NewClient news a basic client
func NewClient(url, authType, accessCredential string, skipCertVerify bool) (Client, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipCertVerify,
		},
	}

	authorizer, err := auth.GetAuthorizer(authType, accessCredential)
	if err != nil {
		return nil, errors.Wrap(err, "new v1 client")
	}

	return &basicClient{
		httpClient: &http.Client{
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		spec:       NewSpec(url),
		authorizer: authorizer,
	}, nil
}

// GetMetadata ...
func (c *basicClient) GetMetadata() (*ScannerAdapterMetadata, error) {
	def := c.spec.Metadata()

	request, err := http.NewRequest(http.MethodGet, def.URL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "v1 client: get metadata")
	}

	// Resolve header
	def.Resolver(request)

	// Send request
	respData, err := c.send(request, generalResponseHandler(http.StatusOK))
	if err != nil {
		return nil, errors.Wrap(err, "v1 client: get metadata")
	}

	// Unmarshal data
	meta := &ScannerAdapterMetadata{}
	if err := json.Unmarshal(respData, meta); err != nil {
		return nil, errors.Wrap(err, "v1 client: get metadata")
	}

	return meta, nil
}

// SubmitScan ...
func (c *basicClient) SubmitScan(req *ScanRequest) (*ScanResponse, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "v1 client: submit scan")
	}

	def := c.spec.SubmitScan()
	request, err := http.NewRequest(http.MethodPost, def.URL, bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, "v1 client: submit scan")
	}

	// Resolve header
	def.Resolver(request)

	respData, err := c.send(request, generalResponseHandler(http.StatusAccepted))
	if err != nil {
		return nil, errors.Wrap(err, "v1 client: submit scan")
	}

	resp := &ScanResponse{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, errors.Wrap(err, "v1 client: submit scan")
	}

	return resp, nil
}

// GetScanReport ...
func (c *basicClient) GetScanReport(scanRequestID, reportMIMEType string) (string, error) {
	if len(scanRequestID) == 0 {
		return "", errors.New("empty scan request ID")
	}

	if len(reportMIMEType) == 0 {
		return "", errors.New("missing report mime type")
	}

	def := c.spec.GetScanReport(scanRequestID, reportMIMEType)

	req, err := http.NewRequest(http.MethodGet, def.URL, nil)
	if err != nil {
		return "", errors.Wrap(err, "v1 client: get scan report")
	}

	// Resolve header
	def.Resolver(req)

	respData, err := c.send(req, reportResponseHandler())
	if err != nil {
		// This error should not be wrapped
		return "", err
	}

	return string(respData), nil
}

func (c *basicClient) send(req *http.Request, h responseHandler) ([]byte, error) {
	if c.authorizer != nil {
		if err := c.authorizer.Authorize(req); err != nil {
			return nil, errors.Wrap(err, "send: authorization")
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Just logged
			logger.Errorf("close response body error: %s", err)
		}
	}()

	return h(resp.StatusCode, resp)
}

// responseHandlerFunc is a handler func template for handling the http response data,
// especially the error part.
type responseHandler func(code int, resp *http.Response) ([]byte, error)

// generalResponseHandler create a general response handler to cover the common cases.
func generalResponseHandler(expectedCode int) responseHandler {
	return func(code int, resp *http.Response) ([]byte, error) {
		return generalRespHandlerFunc(expectedCode, code, resp)
	}
}

// reportResponseHandler creates response handler for get report special case.
func reportResponseHandler() responseHandler {
	return func(code int, resp *http.Response) ([]byte, error) {
		if code == http.StatusFound {
			// Set default
			retryAfter := defaultRefreshInterval // seconds
			// Read `retry after` info from header
			v := resp.Header.Get(refreshAfterHeader)
			if len(v) > 0 {
				if i, err := strconv.ParseInt(v, 10, 8); err == nil {
					retryAfter = int(i)
				} else {
					// log error
					logger.Errorf("Parse `%s` error: %s", refreshAfterHeader, err)
				}
			}

			return nil, &ReportNotReadyError{RetryAfter: retryAfter}
		}

		return generalRespHandlerFunc(http.StatusOK, code, resp)
	}
}

// generalRespHandlerFunc is a handler to cover the general cases
func generalRespHandlerFunc(expectedCode, code int, resp *http.Response) ([]byte, error) {
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if code != expectedCode {
		if len(buf) > 0 {
			// Try to read error response
			eResp := &ErrorResponse{
				Err: &Error{},
			}

			err := json.Unmarshal(buf, eResp)
			if err != nil {
				return nil, errors.Wrap(err, "general response handler")
			}

			// Append more contexts
			eResp.Err.Message = fmt.Sprintf(
				"%s: general response handler: unexpected status code: %d, expected: %d",
				eResp.Err.Message,
				code,
				expectedCode,
			)

			return nil, eResp
		}

		return nil, errors.Errorf("general response handler: unexpected status code: %d, expected: %d", code, expectedCode)
	}

	return buf, nil
}
