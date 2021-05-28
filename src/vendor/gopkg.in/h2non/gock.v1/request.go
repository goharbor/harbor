package gock

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// MapRequestFunc represents the required function interface for request mappers.
type MapRequestFunc func(*http.Request) *http.Request

// FilterRequestFunc represents the required function interface for request filters.
type FilterRequestFunc func(*http.Request) bool

// Request represents the high-level HTTP request used to store
// request fields used to match intercepted requests.
type Request struct {
	// Mock stores the parent mock reference for the current request mock used for method delegation.
	Mock Mock

	// Response stores the current Response instance for the current matches Request.
	Response *Response

	// Error stores the latest mock request configuration error.
	Error error

	// Counter stores the pending times that the current mock should be active.
	Counter int

	// Persisted stores if the current mock should be always active.
	Persisted bool

	// Options stores options for current Request.
	Options Options

	// URLStruct stores the parsed URL as *url.URL struct.
	URLStruct *url.URL

	// Method stores the Request HTTP method to match.
	Method string

	// CompressionScheme stores the Request Compression scheme to match and use for decompression.
	CompressionScheme string

	// Header stores the HTTP header fields to match.
	Header http.Header

	// Cookies stores the Request HTTP cookies values to match.
	Cookies []*http.Cookie

	// PathParams stores the path parameters to match.
	PathParams map[string]string

	// BodyBuffer stores the body data to match.
	BodyBuffer []byte

	// Mappers stores the request functions mappers used for matching.
	Mappers []MapRequestFunc

	// Filters stores the request functions filters used for matching.
	Filters []FilterRequestFunc
}

// NewRequest creates a new Request instance.
func NewRequest() *Request {
	return &Request{
		Counter:    1,
		URLStruct:  &url.URL{},
		Header:     make(http.Header),
		PathParams: make(map[string]string),
	}
}

// URL defines the mock URL to match.
func (r *Request) URL(uri string) *Request {
	r.URLStruct, r.Error = url.Parse(uri)
	return r
}

// SetURL defines the url.URL struct to be used for matching.
func (r *Request) SetURL(u *url.URL) *Request {
	r.URLStruct = u
	return r
}

// Path defines the mock URL path value to match.
func (r *Request) Path(path string) *Request {
	r.URLStruct.Path = path
	return r
}

// Get specifies the GET method and the given URL path to match.
func (r *Request) Get(path string) *Request {
	return r.method("GET", path)
}

// Post specifies the POST method and the given URL path to match.
func (r *Request) Post(path string) *Request {
	return r.method("POST", path)
}

// Put specifies the PUT method and the given URL path to match.
func (r *Request) Put(path string) *Request {
	return r.method("PUT", path)
}

// Delete specifies the DELETE method and the given URL path to match.
func (r *Request) Delete(path string) *Request {
	return r.method("DELETE", path)
}

// Patch specifies the PATCH method and the given URL path to match.
func (r *Request) Patch(path string) *Request {
	return r.method("PATCH", path)
}

// Head specifies the HEAD method and the given URL path to match.
func (r *Request) Head(path string) *Request {
	return r.method("HEAD", path)
}

// method is a DRY shortcut used to declare the expected HTTP method and URL path.
func (r *Request) method(method, path string) *Request {
	if path != "/" {
		r.URLStruct.Path = path
	}
	r.Method = strings.ToUpper(method)
	return r
}

// Body defines the body data to match based on a io.Reader interface.
func (r *Request) Body(body io.Reader) *Request {
	r.BodyBuffer, r.Error = ioutil.ReadAll(body)
	return r
}

// BodyString defines the body to match based on a given string.
func (r *Request) BodyString(body string) *Request {
	r.BodyBuffer = []byte(body)
	return r
}

// File defines the body to match based on the given file path string.
func (r *Request) File(path string) *Request {
	r.BodyBuffer, r.Error = ioutil.ReadFile(path)
	return r
}

// Compression defines the request compression scheme, and enables automatic body decompression.
// Supports only the "gzip" scheme so far.
func (r *Request) Compression(scheme string) *Request {
	r.Header.Set("Content-Encoding", scheme)
	r.CompressionScheme = scheme
	return r
}

// JSON defines the JSON body to match based on a given structure.
func (r *Request) JSON(data interface{}) *Request {
	if r.Header.Get("Content-Type") == "" {
		r.Header.Set("Content-Type", "application/json")
	}
	r.BodyBuffer, r.Error = readAndDecode(data, "json")
	return r
}

// XML defines the XML body to match based on a given structure.
func (r *Request) XML(data interface{}) *Request {
	if r.Header.Get("Content-Type") == "" {
		r.Header.Set("Content-Type", "application/xml")
	}
	r.BodyBuffer, r.Error = readAndDecode(data, "xml")
	return r
}

// MatchType defines the request Content-Type MIME header field.
// Supports type alias. E.g: json, xml, form, text...
func (r *Request) MatchType(kind string) *Request {
	mime := BodyTypeAliases[kind]
	if mime != "" {
		kind = mime
	}
	r.Header.Set("Content-Type", kind)
	return r
}

// BasicAuth defines a username and password for HTTP Basic Authentication
func (r *Request) BasicAuth(username, password string) *Request {
	r.Header.Set("Authorization", "Basic "+basicAuth(username, password))
	return r
}

// MatchHeader defines a new key and value header to match.
func (r *Request) MatchHeader(key, value string) *Request {
	r.Header[key] = []string{value}
	return r
}

// HeaderPresent defines that a header field must be present in the request.
func (r *Request) HeaderPresent(key string) *Request {
	r.Header[key] = []string{".*"}
	return r
}

// MatchHeaders defines a map of key-value headers to match.
func (r *Request) MatchHeaders(headers map[string]string) *Request {
	for key, value := range headers {
		r.Header[key] = []string{value}
	}
	return r
}

// MatchParam defines a new key and value URL query param to match.
func (r *Request) MatchParam(key, value string) *Request {
	query := r.URLStruct.Query()
	query.Set(key, value)
	r.URLStruct.RawQuery = query.Encode()
	return r
}

// MatchParams defines a map of URL query param key-value to match.
func (r *Request) MatchParams(params map[string]string) *Request {
	query := r.URLStruct.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	r.URLStruct.RawQuery = query.Encode()
	return r
}

// ParamPresent matches if the given query param key is present in the URL.
func (r *Request) ParamPresent(key string) *Request {
	r.MatchParam(key, ".*")
	return r
}

// PathParam matches if a given path parameter key is present in the URL.
//
// The value is representative of the restful resource the key defines, e.g.
//   // /users/123/name
//   r.PathParam("users", "123")
// would match.
func (r *Request) PathParam(key, val string) *Request {
	r.PathParams[key] = val

	return r
}

// Persist defines the current HTTP mock as persistent and won't be removed after intercepting it.
func (r *Request) Persist() *Request {
	r.Persisted = true
	return r
}

// WithOptions sets the options for the request.
func (r *Request) WithOptions(options Options) *Request {
	r.Options = options
	return r
}

// Times defines the number of times that the current HTTP mock should remain active.
func (r *Request) Times(num int) *Request {
	r.Counter = num
	return r
}

// AddMatcher adds a new matcher function to match the request.
func (r *Request) AddMatcher(fn MatchFunc) *Request {
	r.Mock.AddMatcher(fn)
	return r
}

// SetMatcher sets a new matcher function to match the request.
func (r *Request) SetMatcher(matcher Matcher) *Request {
	r.Mock.SetMatcher(matcher)
	return r
}

// Map adds a new request mapper function to map http.Request before the matching process.
func (r *Request) Map(fn MapRequestFunc) *Request {
	r.Mappers = append(r.Mappers, fn)
	return r
}

// Filter filters a new request filter function to filter http.Request before the matching process.
func (r *Request) Filter(fn FilterRequestFunc) *Request {
	r.Filters = append(r.Filters, fn)
	return r
}

// EnableNetworking enables the use real networking for the current mock.
func (r *Request) EnableNetworking() *Request {
	if r.Response != nil {
		r.Response.UseNetwork = true
	}
	return r
}

// Reply defines the Response status code and returns the mock Response DSL.
func (r *Request) Reply(status int) *Response {
	return r.Response.Status(status)
}

// ReplyError defines the Response simulated error.
func (r *Request) ReplyError(err error) *Response {
	return r.Response.SetError(err)
}

// ReplyFunc allows the developer to define the mock response via a custom function.
func (r *Request) ReplyFunc(replier func(*Response)) *Response {
	replier(r.Response)
	return r.Response
}

// See 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
