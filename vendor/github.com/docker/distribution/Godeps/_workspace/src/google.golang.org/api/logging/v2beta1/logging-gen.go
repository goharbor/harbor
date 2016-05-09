// Package logging provides access to the Google Cloud Logging API.
//
// See https://cloud.google.com/logging/docs/
//
// Usage example:
//
//   import "google.golang.org/api/logging/v2beta1"
//   ...
//   loggingService, err := logging.New(oauthHttpClient)
package logging // import "google.golang.org/api/logging/v2beta1"

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

const apiId = "logging:v2beta1"
const apiName = "logging"
const apiVersion = "v2beta1"
const basePath = "https://logging.googleapis.com/"

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

// LogLine: Application log line emitted while processing a request.
type LogLine struct {
	// LogMessage: App provided log message.
	LogMessage string `json:"logMessage,omitempty"`

	// Severity: Severity of log.
	//
	// Possible values:
	//   "DEFAULT"
	//   "DEBUG"
	//   "INFO"
	//   "NOTICE"
	//   "WARNING"
	//   "ERROR"
	//   "CRITICAL"
	//   "ALERT"
	//   "EMERGENCY"
	Severity string `json:"severity,omitempty"`

	// SourceLocation: Line of code that generated this log message.
	SourceLocation *SourceLocation `json:"sourceLocation,omitempty"`

	// Time: Time when log entry was made. May be inaccurate.
	Time string `json:"time,omitempty"`

	// ForceSendFields is a list of field names (e.g. "LogMessage") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *LogLine) MarshalJSON() ([]byte, error) {
	type noMethod LogLine
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// RequestLog: Complete log information about a single request to an
// application.
type RequestLog struct {
	// AppEngineRelease: App Engine release version string.
	AppEngineRelease string `json:"appEngineRelease,omitempty"`

	// AppId: Identifies the application that handled this request.
	AppId string `json:"appId,omitempty"`

	// Cost: An indication of the relative cost of serving this request.
	Cost float64 `json:"cost,omitempty"`

	// EndTime: Time at which request was known to end processing.
	EndTime string `json:"endTime,omitempty"`

	// Finished: If true, represents a finished request. Otherwise, the
	// request is active.
	Finished bool `json:"finished,omitempty"`

	// Host: The Internet host and port number of the resource being
	// requested.
	Host string `json:"host,omitempty"`

	// HttpVersion: HTTP version of request.
	HttpVersion string `json:"httpVersion,omitempty"`

	// InstanceId: An opaque identifier for the instance that handled the
	// request.
	InstanceId string `json:"instanceId,omitempty"`

	// InstanceIndex: If the instance that processed this request was
	// individually addressable (i.e. belongs to a manually scaled module),
	// this is the index of the instance.
	InstanceIndex int64 `json:"instanceIndex,omitempty"`

	// Ip: Origin IP address.
	Ip string `json:"ip,omitempty"`

	// Latency: Latency of the request.
	Latency string `json:"latency,omitempty"`

	// Line: List of log lines emitted by the application while serving this
	// request, if requested.
	Line []*LogLine `json:"line,omitempty"`

	// MegaCycles: Number of CPU megacycles used to process request.
	MegaCycles int64 `json:"megaCycles,omitempty,string"`

	// Method: Request method, such as `GET`, `HEAD`, `PUT`, `POST`, or
	// `DELETE`.
	Method string `json:"method,omitempty"`

	// ModuleId: Identifies the module of the application that handled this
	// request.
	ModuleId string `json:"moduleId,omitempty"`

	// Nickname: A string that identifies a logged-in user who made this
	// request, or empty if the user is not logged in. Most likely, this is
	// the part of the user's email before the '@' sign. The field value is
	// the same for different requests from the same user, but different
	// users may have a similar name. This information is also available to
	// the application via Users API. This field will be populated starting
	// with App Engine 1.9.21.
	Nickname string `json:"nickname,omitempty"`

	// PendingTime: Time this request spent in the pending request queue, if
	// it was pending at all.
	PendingTime string `json:"pendingTime,omitempty"`

	// Referrer: Referrer URL of request.
	Referrer string `json:"referrer,omitempty"`

	// RequestId: Globally unique identifier for a request, based on request
	// start time. Request IDs for requests which started later will compare
	// greater as binary strings than those for requests which started
	// earlier.
	RequestId string `json:"requestId,omitempty"`

	// Resource: Contains the path and query portion of the URL that was
	// requested. For example, if the URL was
	// "http://example.com/app?name=val", the resource would be
	// "/app?name=val". Any trailing fragment (separated by a '#' character)
	// will not be included.
	Resource string `json:"resource,omitempty"`

	// ResponseSize: Size in bytes sent back to client by request.
	ResponseSize int64 `json:"responseSize,omitempty,string"`

	// SourceReference: Source code for the application that handled this
	// request. There can be more than one source reference per deployed
	// application if source code is distributed among multiple
	// repositories.
	SourceReference []*SourceReference `json:"sourceReference,omitempty"`

	// StartTime: Time at which request was known to have begun processing.
	StartTime string `json:"startTime,omitempty"`

	// Status: Response status of request.
	Status int64 `json:"status,omitempty"`

	// TaskName: Task name of the request (for an offline request).
	TaskName string `json:"taskName,omitempty"`

	// TaskQueueName: Queue name of the request (for an offline request).
	TaskQueueName string `json:"taskQueueName,omitempty"`

	// TraceId: Cloud Trace identifier of the trace for this request.
	TraceId string `json:"traceId,omitempty"`

	// UrlMapEntry: File or class within URL mapping used for request.
	// Useful for tracking down the source code which was responsible for
	// managing request. Especially for multiply mapped handlers.
	UrlMapEntry string `json:"urlMapEntry,omitempty"`

	// UserAgent: User agent used for making request.
	UserAgent string `json:"userAgent,omitempty"`

	// VersionId: Version of the application that handled this request.
	VersionId string `json:"versionId,omitempty"`

	// WasLoadingRequest: Was this request a loading request for this
	// instance?
	WasLoadingRequest bool `json:"wasLoadingRequest,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AppEngineRelease") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *RequestLog) MarshalJSON() ([]byte, error) {
	type noMethod RequestLog
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// SourceLocation: Specifies a location in a source file.
type SourceLocation struct {
	// File: Source file name. May or may not be a fully qualified name,
	// depending on the runtime environment.
	File string `json:"file,omitempty"`

	// FunctionName: Human-readable name of the function or method being
	// invoked, with optional context such as the class or package name, for
	// use in contexts such as the logs viewer where file:line number is
	// less meaningful. This may vary by language, for example: in Java:
	// qual.if.ied.Class.method in Go: dir/package.func in Python: function
	// ...
	FunctionName string `json:"functionName,omitempty"`

	// Line: Line within the source file.
	Line int64 `json:"line,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "File") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SourceLocation) MarshalJSON() ([]byte, error) {
	type noMethod SourceLocation
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}

// SourceReference: A reference to a particular snapshot of the source
// tree used to build and deploy an application.
type SourceReference struct {
	// Repository: Optional. A URI string identifying the repository.
	// Example: "https://github.com/GoogleCloudPlatform/kubernetes.git"
	Repository string `json:"repository,omitempty"`

	// RevisionId: The canonical (and persistent) identifier of the deployed
	// revision. Example (git): "0035781c50ec7aa23385dc841529ce8a4b70db1b"
	RevisionId string `json:"revisionId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Repository") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SourceReference) MarshalJSON() ([]byte, error) {
	type noMethod SourceReference
	raw := noMethod(*s)
	return internal.MarshalJSON(raw, s.ForceSendFields)
}
