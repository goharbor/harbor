package chartserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common"
	hlog "github.com/goharbor/harbor/src/common/utils/log"
	n_event "github.com/goharbor/harbor/src/core/notifier/event"
	"github.com/goharbor/harbor/src/replication"
	rep_event "github.com/goharbor/harbor/src/replication/event"
	"github.com/justinas/alice"
	"time"
)

const (
	agentHarbor         = "HARBOR"
	contentLengthHeader = "Content-Length"

	defaultRepo             = "library"
	rootUploadingEndpoint   = "/api/chartrepo/charts"
	rootIndexEndpoint       = "/chartrepo/index.yaml"
	chartRepoHealthEndpoint = "/api/chartrepo/health"
)

// ProxyEngine is used to proxy the related traffics
type ProxyEngine struct {
	// The backend target server the traffic will be forwarded to
	// Just in case we'll use it
	backend *url.URL

	// Use go reverse proxy as engine
	engine http.Handler
}

// NewProxyEngine is constructor of NewProxyEngine
func NewProxyEngine(target *url.URL, cred *Credential, chains ...*alice.Chain) *ProxyEngine {
	var engine http.Handler

	engine = &httputil.ReverseProxy{
		ErrorLog: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
		Director: func(req *http.Request) {
			director(target, cred, req)
		},
		ModifyResponse: modifyResponse,
	}

	if len(chains) > 0 {
		hlog.Info("New chart server traffic proxy with middlewares")
		engine = chains[0].Then(engine)
	}

	return &ProxyEngine{
		backend: target,
		engine:  engine,
	}
}

// ServeHTTP serves the incoming http requests
func (pe *ProxyEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pe.engine.ServeHTTP(w, req)
}

// Overwrite the http requests
func director(target *url.URL, cred *Credential, req *http.Request) {
	// Closure
	targetQuery := target.RawQuery

	// Overwrite the request URL to the target path
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	rewriteURLPath(req)
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", agentHarbor)
	}

	// Add authentication header if it is existing
	if cred != nil {
		req.SetBasicAuth(cred.Username, cred.Password)
	}
}

// Modify the http response
func modifyResponse(res *http.Response) error {
	// Upload chart success, then to the notification to replication handler
	if res.StatusCode == http.StatusCreated {
		// 201 and has chart_upload_event context
		// means this response is for uploading chart success.
		chartUploadEvent := res.Request.Context().Value(common.ChartUploadCtxKey)
		e, ok := chartUploadEvent.(*rep_event.Event)
		if !ok {
			hlog.Error("failed to convert chart upload context into replication event.")
		} else {
			// Todo: it used as the replacement of webhook, will be removed when webhook to be introduced.
			go func() {
				if err := replication.EventHandler.Handle(e); err != nil {
					hlog.Errorf("failed to handle event: %v", err)
				}
			}()

			// Trigger harbor webhook
			if e != nil && e.Resource != nil && e.Resource.Metadata != nil && len(e.Resource.Metadata.Vtags) > 0 &&
				len(e.Resource.ExtendedInfo) > 0 {
				event := &n_event.Event{}
				metaData := &n_event.ChartUploadMetaData{
					ChartMetaData: n_event.ChartMetaData{
						ProjectName: e.Resource.ExtendedInfo["projectName"].(string),
						ChartName:   e.Resource.ExtendedInfo["chartName"].(string),
						Versions:    e.Resource.Metadata.Vtags,
						OccurAt:     time.Now(),
						Operator:    e.Resource.ExtendedInfo["operator"].(string),
					},
				}
				if err := event.Build(metaData); err == nil {
					if err := event.Publish(); err != nil {
						hlog.Errorf("failed to publish chart upload event: %v", err)
					}
				} else {
					hlog.Errorf("failed to build chart upload event metadata: %v", err)
				}
			}
		}
	}

	// Process downloading chart success webhook event
	if res.StatusCode == http.StatusOK {
		chartDownloadEvent := res.Request.Context().Value(common.ChartDownloadCtxKey)
		eventMetaData, ok := chartDownloadEvent.(*n_event.ChartDownloadMetaData)
		if ok && eventMetaData != nil {
			// Trigger harbor webhook
			event := &n_event.Event{}
			if err := event.Build(eventMetaData); err == nil {
				if err := event.Publish(); err != nil {
					hlog.Errorf("failed to publish chart download event: %v", err)
				}
			} else {
				hlog.Errorf("failed to build chart download event metadata: %v", err)
			}
		}
	}

	// Accept cases
	// Success or redirect
	if res.StatusCode >= http.StatusOK && res.StatusCode <= http.StatusTemporaryRedirect {
		return nil
	}

	// Detect the 401 code, if it is,overwrite it to 500.
	// We also re-write the error content to structural error object
	errorObj := make(map[string]string)
	if res.StatusCode == http.StatusUnauthorized {
		errorObj["error"] = "operation request from unauthorized source is rejected"
		res.StatusCode = http.StatusInternalServerError
	} else {
		// Extract the error and wrap it into the error object
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			errorObj["error"] = fmt.Sprintf("%s: %s", res.Status, err.Error())
		} else {
			if err := json.Unmarshal(data, &errorObj); err != nil {
				errorObj["error"] = string(data)
			}
		}
	}

	content, err := json.Marshal(errorObj)
	if err != nil {
		return err
	}

	size := len(content)
	body := ioutil.NopCloser(bytes.NewReader(content))
	res.Body = body
	res.ContentLength = int64(size)
	res.Header.Set(contentLengthHeader, strconv.Itoa(size))

	return nil
}

// Join the path
// Copy from the go reverse proxy
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// Rewrite the incoming URL with the right backend URL pattern
// Remove 'chartrepo' from the endpoints of manipulation API
// Remove 'chartrepo' from the endpoints of repository services
func rewriteURLPath(req *http.Request) {
	incomingURLPath := req.URL.Path

	// Health check endpoint
	if incomingURLPath == chartRepoHealthEndpoint {
		req.URL.Path = "/health"
		return
	}

	// Root uploading endpoint
	if incomingURLPath == rootUploadingEndpoint {
		req.URL.Path = strings.Replace(incomingURLPath, "chartrepo", defaultRepo, 1)
		return
	}

	// Repository endpoints
	if strings.HasPrefix(incomingURLPath, "/chartrepo") {
		req.URL.Path = strings.TrimPrefix(incomingURLPath, "/chartrepo")
		return
	}

	// API endpoints
	if strings.HasPrefix(incomingURLPath, "/api/chartrepo") {
		req.URL.Path = strings.Replace(incomingURLPath, "/chartrepo", "", 1)
		return
	}
}
