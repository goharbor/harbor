// Package cloudtrace provides access to the Google Cloud Trace API.
//
// See https://cloud.google.com/tools/cloud-trace
//
// Usage example:
//
//   import "google.golang.org/api/cloudtrace/v1"
//   ...
//   cloudtraceService, err := cloudtrace.New(oauthHttpClient)
package cloudtrace // import "google.golang.org/api/cloudtrace/v1"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/internal"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace
var _ = internal.MarshalJSON
var _ = context.Canceled
var _ = ctxhttp.Do

const apiId = "cloudtrace:v1"
const apiName = "cloudtrace"
const apiVersion = "v1"
const basePath = "https://cloudtrace.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// View and manage your data across Google Cloud Platform services
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Projects = NewProjectsService(s)
	s.V1 = NewV1Service(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Projects *ProjectsService

	V1 *V1Service
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewProjectsService(s *Service) *ProjectsService {
	rs := &ProjectsService{s: s}
	rs.Traces = NewProjectsTracesService(s)
	return rs
}

type ProjectsService struct {
	s *Service

	Traces *ProjectsTracesService
}

func NewProjectsTracesService(s *Service) *ProjectsTracesService {
	rs := &ProjectsTracesService{s: s}
	return rs
}

type ProjectsTracesService struct {
	s *Service
}

func NewV1Service(s *Service) *V1Service {
	rs := &V1Service{s: s}
	return rs
}

type V1Service struct {
	s *Service
}

// Empty: A generic empty message that you can re-use to avoid defining
// duplicated empty messages in your APIs. A typical example is to use
// it as the request or the response type of an API method. For
// instance: service Foo { rpc Bar(google.protobuf.Empty) returns
// (google.protobuf.Empty); } The JSON representation for `Empty` is
// empty JSON object `{}`.
type Empty struct {
	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`
}

// ListTracesResponse: The response message for the ListTraces method.
type ListTracesResponse struct {
	// NextPageToken: If defined, indicates that there are more topics that
	// match the request, and this value should be passed to the next
	// ListTopicsRequest to continue.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// Traces: The list of trace records returned.
	Traces []*Trace `json:"traces,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "NextPageToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListTracesResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListTracesResponse
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// Trace: A Trace is a collection of spans describing the execution
// timings of a single operation.
type Trace struct {
	// ProjectId: The Project ID of the Google Cloud project.
	ProjectId string `json:"projectId,omitempty"`

	// Spans: The collection of span records within this trace. Spans that
	// appear in calls to PatchTraces may be incomplete or partial.
	Spans []*TraceSpan `json:"spans,omitempty"`

	// TraceId: A 128-bit numeric value, formatted as a 32-byte hex string,
	// that represents a trace. Each trace should have an identifier that is
	// globally unique.
	TraceId string `json:"traceId,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ProjectId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Trace) MarshalJSON() ([]byte, error) {
	type noMethod Trace
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// TraceSpan: A span is the data recorded with a single span.
type TraceSpan struct {
	// EndTime: The end time of the span in nanoseconds from the UNIX epoch.
	EndTime string `json:"endTime,omitempty"`

	// Kind: SpanKind distinguishes spans generated in a particular context.
	// For example, two spans with the same name, one with the kind
	// RPC_CLIENT, and the other with RPC_SERVER can indicate the queueing
	// latency associated with the span.
	//
	// Possible values:
	//   "SPAN_KIND_UNSPECIFIED"
	//   "RPC_SERVER"
	//   "RPC_CLIENT"
	Kind string `json:"kind,omitempty"`

	// Labels: Annotations via labels.
	Labels map[string]string `json:"labels,omitempty"`

	// Name: The name of the trace. This is sanitized and displayed on the
	// UI. This may be a method name or some other per-callsite name. For
	// the same binary and the same call point, it is a good practice to
	// choose a consistent name in order to correlate cross-trace spans.
	Name string `json:"name,omitempty"`

	// ParentSpanId: Identifies the parent of the current span. May be
	// missing. Serialized bytes representation of SpanId.
	ParentSpanId uint64 `json:"parentSpanId,omitempty,string"`

	// SpanId: Identifier of the span within the trace. Each span should
	// have an identifier that is unique per trace.
	SpanId uint64 `json:"spanId,omitempty,string"`

	// StartTime: The start time of the span in nanoseconds from the UNIX
	// epoch.
	StartTime string `json:"startTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "EndTime") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TraceSpan) MarshalJSON() ([]byte, error) {
	type noMethod TraceSpan
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// Traces: A list of traces for the PatchTraces method.
type Traces struct {
	// Traces: A list of traces.
	Traces []*Trace `json:"traces,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Traces") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Traces) MarshalJSON() ([]byte, error) {
	type noMethod Traces
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// method id "cloudtrace.projects.patchTraces":

type ProjectsPatchTracesCall struct {
	s         *Service
	projectId string
	traces    *Traces
	opt_      map[string]interface{}
	ctx_      context.Context
}

// PatchTraces: Updates the existing traces specified by
// PatchTracesRequest and inserts the new traces. Any existing trace or
// span fields included in an update are overwritten by the update, and
// any additional fields in an update are merged with the existing trace
// data.
func (r *ProjectsService) PatchTraces(projectId string, traces *Traces) *ProjectsPatchTracesCall {
	c := &ProjectsPatchTracesCall{s: r.s, opt_: make(map[string]interface{})}
	c.projectId = projectId
	c.traces = traces
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsPatchTracesCall) Fields(s ...googleapi.Field) *ProjectsPatchTracesCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ProjectsPatchTracesCall) Context(ctx context.Context) *ProjectsPatchTracesCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsPatchTracesCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.traces)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/traces")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("PATCH", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "cloudtrace.projects.patchTraces" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsPatchTracesCall) Do() (*Empty, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Empty{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates the existing traces specified by PatchTracesRequest and inserts the new traces. Any existing trace or span fields included in an update are overwritten by the update, and any additional fields in an update are merged with the existing trace data.",
	//   "httpMethod": "PATCH",
	//   "id": "cloudtrace.projects.patchTraces",
	//   "parameterOrder": [
	//     "projectId"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The project id of the trace to patch.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/traces",
	//   "request": {
	//     "$ref": "Traces"
	//   },
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "cloudtrace.projects.traces.get":

type ProjectsTracesGetCall struct {
	s         *Service
	projectId string
	traceId   string
	opt_      map[string]interface{}
	ctx_      context.Context
}

// Get: Gets one trace by id.
func (r *ProjectsTracesService) Get(projectId string, traceId string) *ProjectsTracesGetCall {
	c := &ProjectsTracesGetCall{s: r.s, opt_: make(map[string]interface{})}
	c.projectId = projectId
	c.traceId = traceId
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsTracesGetCall) Fields(s ...googleapi.Field) *ProjectsTracesGetCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsTracesGetCall) IfNoneMatch(entityTag string) *ProjectsTracesGetCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ProjectsTracesGetCall) Context(ctx context.Context) *ProjectsTracesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsTracesGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/traces/{traceId}")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"traceId":   c.traceId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "cloudtrace.projects.traces.get" call.
// Exactly one of *Trace or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Trace.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsTracesGetCall) Do() (*Trace, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Trace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets one trace by id.",
	//   "httpMethod": "GET",
	//   "id": "cloudtrace.projects.traces.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "traceId"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The project id of the trace to return.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "traceId": {
	//       "description": "The trace id of the trace to return.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/traces/{traceId}",
	//   "response": {
	//     "$ref": "Trace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "cloudtrace.projects.traces.list":

type ProjectsTracesListCall struct {
	s         *Service
	projectId string
	opt_      map[string]interface{}
	ctx_      context.Context
}

// List: List traces matching the filter expression.
func (r *ProjectsTracesService) List(projectId string) *ProjectsTracesListCall {
	c := &ProjectsTracesListCall{s: r.s, opt_: make(map[string]interface{})}
	c.projectId = projectId
	return c
}

// EndTime sets the optional parameter "endTime": Start of the time
// interval (exclusive).
func (c *ProjectsTracesListCall) EndTime(endTime string) *ProjectsTracesListCall {
	c.opt_["endTime"] = endTime
	return c
}

// Filter sets the optional parameter "filter": An optional filter for
// the request.
func (c *ProjectsTracesListCall) Filter(filter string) *ProjectsTracesListCall {
	c.opt_["filter"] = filter
	return c
}

// OrderBy sets the optional parameter "orderBy": The trace field used
// to establish the order of traces returned by the ListTraces method.
// Possible options are: trace_id name (name field of root span)
// duration (different between end_time and start_time fields of root
// span) start (start_time field of root span) Descending order can be
// specified by appending "desc" to the sort field: name desc Only one
// sort field is permitted, though this may change in the future.
func (c *ProjectsTracesListCall) OrderBy(orderBy string) *ProjectsTracesListCall {
	c.opt_["orderBy"] = orderBy
	return c
}

// PageSize sets the optional parameter "pageSize": Maximum number of
// topics to return. If not specified or <= 0, the implementation will
// select a reasonable value. The implemenation may always return fewer
// than the requested page_size.
func (c *ProjectsTracesListCall) PageSize(pageSize int64) *ProjectsTracesListCall {
	c.opt_["pageSize"] = pageSize
	return c
}

// PageToken sets the optional parameter "pageToken": The token
// identifying the page of results to return from the ListTraces method.
// If present, this value is should be taken from the next_page_token
// field of a previous ListTracesResponse.
func (c *ProjectsTracesListCall) PageToken(pageToken string) *ProjectsTracesListCall {
	c.opt_["pageToken"] = pageToken
	return c
}

// StartTime sets the optional parameter "startTime": End of the time
// interval (inclusive).
func (c *ProjectsTracesListCall) StartTime(startTime string) *ProjectsTracesListCall {
	c.opt_["startTime"] = startTime
	return c
}

// View sets the optional parameter "view": ViewType specifies the
// projection of the result.
//
// Possible values:
//   "VIEW_TYPE_UNSPECIFIED"
//   "MINIMAL"
//   "ROOTSPAN"
//   "COMPLETE"
func (c *ProjectsTracesListCall) View(view string) *ProjectsTracesListCall {
	c.opt_["view"] = view
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsTracesListCall) Fields(s ...googleapi.Field) *ProjectsTracesListCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsTracesListCall) IfNoneMatch(entityTag string) *ProjectsTracesListCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *ProjectsTracesListCall) Context(ctx context.Context) *ProjectsTracesListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsTracesListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["endTime"]; ok {
		params.Set("endTime", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["filter"]; ok {
		params.Set("filter", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["orderBy"]; ok {
		params.Set("orderBy", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageSize"]; ok {
		params.Set("pageSize", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["pageToken"]; ok {
		params.Set("pageToken", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["startTime"]; ok {
		params.Set("startTime", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["view"]; ok {
		params.Set("view", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/traces")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "cloudtrace.projects.traces.list" call.
// Exactly one of *ListTracesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListTracesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsTracesListCall) Do() (*ListTracesResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListTracesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "List traces matching the filter expression.",
	//   "httpMethod": "GET",
	//   "id": "cloudtrace.projects.traces.list",
	//   "parameterOrder": [
	//     "projectId"
	//   ],
	//   "parameters": {
	//     "endTime": {
	//       "description": "Start of the time interval (exclusive).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "filter": {
	//       "description": "An optional filter for the request.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "orderBy": {
	//       "description": "The trace field used to establish the order of traces returned by the ListTraces method. Possible options are: trace_id name (name field of root span) duration (different between end_time and start_time fields of root span) start (start_time field of root span) Descending order can be specified by appending \"desc\" to the sort field: name desc Only one sort field is permitted, though this may change in the future.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "Maximum number of topics to return. If not specified or \u003c= 0, the implementation will select a reasonable value. The implemenation may always return fewer than the requested page_size.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The token identifying the page of results to return from the ListTraces method. If present, this value is should be taken from the next_page_token field of a previous ListTracesResponse.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The stringified-version of the project id.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "startTime": {
	//       "description": "End of the time interval (inclusive).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "view": {
	//       "description": "ViewType specifies the projection of the result.",
	//       "enum": [
	//         "VIEW_TYPE_UNSPECIFIED",
	//         "MINIMAL",
	//         "ROOTSPAN",
	//         "COMPLETE"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/traces",
	//   "response": {
	//     "$ref": "ListTracesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "cloudtrace.getDiscovery":

type V1GetDiscoveryCall struct {
	s    *Service
	opt_ map[string]interface{}
	ctx_ context.Context
}

// GetDiscovery: Returns a discovery document in the specified `format`.
// The typeurl in the returned google.protobuf.Any value depends on the
// requested format.
func (r *V1Service) GetDiscovery() *V1GetDiscoveryCall {
	c := &V1GetDiscoveryCall{s: r.s, opt_: make(map[string]interface{})}
	return c
}

// Args sets the optional parameter "args": Any additional arguments.
func (c *V1GetDiscoveryCall) Args(args string) *V1GetDiscoveryCall {
	c.opt_["args"] = args
	return c
}

// Format sets the optional parameter "format": The format requested for
// discovery.
func (c *V1GetDiscoveryCall) Format(format string) *V1GetDiscoveryCall {
	c.opt_["format"] = format
	return c
}

// Labels sets the optional parameter "labels": A list of labels (like
// visibility) influencing the scope of the requested doc.
func (c *V1GetDiscoveryCall) Labels(labels string) *V1GetDiscoveryCall {
	c.opt_["labels"] = labels
	return c
}

// Version sets the optional parameter "version": The API version of the
// requested discovery doc.
func (c *V1GetDiscoveryCall) Version(version string) *V1GetDiscoveryCall {
	c.opt_["version"] = version
	return c
}

// Fields allows partial responses to be retrieved.
// See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *V1GetDiscoveryCall) Fields(s ...googleapi.Field) *V1GetDiscoveryCall {
	c.opt_["fields"] = googleapi.CombineFields(s)
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *V1GetDiscoveryCall) IfNoneMatch(entityTag string) *V1GetDiscoveryCall {
	c.opt_["ifNoneMatch"] = entityTag
	return c
}

// Context sets the context to be used in this call's Do method.
// Any pending HTTP request will be aborted if the provided context
// is canceled.
func (c *V1GetDiscoveryCall) Context(ctx context.Context) *V1GetDiscoveryCall {
	c.ctx_ = ctx
	return c
}

func (c *V1GetDiscoveryCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	params := make(url.Values)
	params.Set("alt", alt)
	if v, ok := c.opt_["args"]; ok {
		params.Set("args", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["format"]; ok {
		params.Set("format", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["labels"]; ok {
		params.Set("labels", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["version"]; ok {
		params.Set("version", fmt.Sprintf("%v", v))
	}
	if v, ok := c.opt_["fields"]; ok {
		params.Set("fields", fmt.Sprintf("%v", v))
	}
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/discovery")
	urls += "?" + params.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if v, ok := c.opt_["ifNoneMatch"]; ok {
		req.Header.Set("If-None-Match", fmt.Sprintf("%v", v))
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "cloudtrace.getDiscovery" call.
func (c *V1GetDiscoveryCall) Do() error {
	res, err := c.doRequest("json")
	if err != nil {
		return err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return err
	}
	return nil
	// {
	//   "description": "Returns a discovery document in the specified `format`. The typeurl in the returned google.protobuf.Any value depends on the requested format.",
	//   "httpMethod": "GET",
	//   "id": "cloudtrace.getDiscovery",
	//   "parameters": {
	//     "args": {
	//       "description": "Any additional arguments.",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     },
	//     "format": {
	//       "description": "The format requested for discovery.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "labels": {
	//       "description": "A list of labels (like visibility) influencing the scope of the requested doc.",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     },
	//     "version": {
	//       "description": "The API version of the requested discovery doc.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/discovery"
	// }

}
