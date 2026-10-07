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
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scan/rest/auth"
)

const (
	// defaultRefreshInterval is the default interval with seconds of refreshing report
	defaultRefreshInterval = 5
	// refreshAfterHeader provides the refresh interval value
	refreshAfterHeader = "Refresh-After"
)

// Client defines the methods to access the optimizer adapter services that
// implement the REST API spec
type Client interface {
	// GetMetadata gets the metadata of the given optimizer adapter
	GetMetadata() (*OptimizerAdapterMetadata, error)

	// SubmitOptimize initiates the optimization of the given artifact.
	//
	//   Returns:
	//     *OptimizeResponse : response with UUID for tracking the result
	//     error             : non nil error if any errors occurred
	SubmitOptimize(req *OptimizeRequest) (*OptimizeResponse, error)

	// GetOptimizationReport gets the optimization result for the corresponding
	// request identifier. Returns *ReportNotReadyError while the adapter is still
	// working (HTTP 302 + Refresh-After header).
	GetOptimizationReport(requestID string) (string, error)
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
		Proxy:        http.ProxyFromEnvironment,
		MaxIdleConns: 100,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipCertVerify,
		},
	}

	authorizer, err := auth.GetAuthorizer(authType, accessCredential)
	if err != nil {
		return nil, errors.Wrap(err, "new optimizer v1 client")
	}

	return &basicClient{
		httpClient: &http.Client{
			Timeout:   time.Second * 5,
			Transport: transport,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		spec:       NewSpec(url),
		authorizer: authorizer,
	}, nil
}

// GetMetadata ...
func (c *basicClient) GetMetadata() (*OptimizerAdapterMetadata, error) {
	def := c.spec.Metadata()

	request, err := http.NewRequest(http.MethodGet, def.URL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "optimizer v1 client: get metadata")
	}

	def.Resolver(request)

	respData, err := c.send(request, generalResponseHandler(http.StatusOK))
	if err != nil {
		return nil, errors.Wrap(err, "optimizer v1 client: get metadata")
	}

	meta := &OptimizerAdapterMetadata{}
	if err := json.Unmarshal(respData, meta); err != nil {
		return nil, errors.Wrap(err, "optimizer v1 client: get metadata")
	}

	return meta, nil
}

// SubmitOptimize ...
func (c *basicClient) SubmitOptimize(req *OptimizeRequest) (*OptimizeResponse, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "optimizer v1 client: submit optimize")
	}

	def := c.spec.SubmitOptimize()
	request, err := http.NewRequest(http.MethodPost, def.URL, bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, "optimizer v1 client: submit optimize")
	}

	def.Resolver(request)

	respData, err := c.send(request, generalResponseHandler(http.StatusAccepted))
	if err != nil {
		return nil, errors.Wrap(err, "optimizer v1 client: submit optimize")
	}

	resp := &OptimizeResponse{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, errors.Wrap(err, "optimizer v1 client: submit optimize")
	}

	return resp, nil
}

// GetOptimizationReport ...
func (c *basicClient) GetOptimizationReport(requestID string) (string, error) {
	if len(requestID) == 0 {
		return "", errors.New("empty optimize request ID")
	}

	def := c.spec.GetOptimizationReport(requestID)
	req, err := http.NewRequest(http.MethodGet, def.URL, nil)
	if err != nil {
		return "", errors.Wrap(err, "optimizer v1 client: get optimization report")
	}

	def.Resolver(req)

	respData, err := c.send(req, reportResponseHandler())
	if err != nil {
		// This error should not be wrapped: callers type-assert *ReportNotReadyError
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
			log.Errorf("close response body error: %s", err)
		}
	}()

	return h(resp.StatusCode, resp)
}

// responseHandler is a handler func template for handling the http response data,
// especially the error part.
type responseHandler func(code int, resp *http.Response) ([]byte, error)

// generalResponseHandler create a general response handler to cover the common cases.
func generalResponseHandler(expectedCode int) responseHandler {
	return func(code int, resp *http.Response) ([]byte, error) {
		return generalRespHandlerFunc(expectedCode, code, resp)
	}
}

// reportResponseHandler creates response handler for the get report special case.
func reportResponseHandler() responseHandler {
	return func(code int, resp *http.Response) ([]byte, error) {
		if code == http.StatusFound {
			retryAfter := defaultRefreshInterval // seconds
			v := resp.Header.Get(refreshAfterHeader)
			if len(v) > 0 {
				if i, err := strconv.ParseInt(v, 10, 8); err == nil {
					retryAfter = int(i)
				} else {
					log.Errorf("Parse `%s` error: %s", refreshAfterHeader, err)
				}
			}

			return nil, &ReportNotReadyError{RetryAfter: retryAfter}
		}

		return generalRespHandlerFunc(http.StatusOK, code, resp)
	}
}

// generalRespHandlerFunc is a handler to cover the general cases
func generalRespHandlerFunc(expectedCode, code int, resp *http.Response) ([]byte, error) {
	buf, err := io.ReadAll(resp.Body)
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
