package gock

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// MapResponseFunc represents the required function interface impletemed by response mappers.
type MapResponseFunc func(*http.Response) *http.Response

// FilterResponseFunc represents the required function interface impletemed by response filters.
type FilterResponseFunc func(*http.Response) bool

// Response represents high-level HTTP fields to configure
// and define HTTP responses intercepted by gock.
type Response struct {
	// Mock stores the parent mock reference for the current response mock used for method delegation.
	Mock Mock

	// Error stores the latest response configuration or injected error.
	Error error

	// UseNetwork enables the use of real network for the current mock.
	UseNetwork bool

	// StatusCode stores the response status code.
	StatusCode int

	// Headers stores the response headers.
	Header http.Header

	// Cookies stores the response cookie fields.
	Cookies []*http.Cookie

	// BodyBuffer stores the array of bytes to use as body.
	BodyBuffer []byte

	// ResponseDelay stores the simulated response delay.
	ResponseDelay time.Duration

	// Mappers stores the request functions mappers used for matching.
	Mappers []MapResponseFunc

	// Filters stores the request functions filters used for matching.
	Filters []FilterResponseFunc
}

// NewResponse creates a new Response.
func NewResponse() *Response {
	return &Response{Header: make(http.Header)}
}

// Status defines the desired HTTP status code to reply in the current response.
func (r *Response) Status(code int) *Response {
	r.StatusCode = code
	return r
}

// Type defines the response Content-Type MIME header field.
// Supports type alias. E.g: json, xml, form, text...
func (r *Response) Type(kind string) *Response {
	mime := BodyTypeAliases[kind]
	if mime != "" {
		kind = mime
	}
	r.Header.Set("Content-Type", kind)
	return r
}

// SetHeader sets a new header field in the mock response.
func (r *Response) SetHeader(key, value string) *Response {
	r.Header.Set(key, value)
	return r
}

// AddHeader adds a new header field in the mock response
// with out removing an existent one.
func (r *Response) AddHeader(key, value string) *Response {
	r.Header.Add(key, value)
	return r
}

// SetHeaders sets a map of header fields in the mock response.
func (r *Response) SetHeaders(headers map[string]string) *Response {
	for key, value := range headers {
		r.Header.Add(key, value)
	}
	return r
}

// Body sets the HTTP response body to be used.
func (r *Response) Body(body io.Reader) *Response {
	r.BodyBuffer, r.Error = ioutil.ReadAll(body)
	return r
}

// BodyString defines the response body as string.
func (r *Response) BodyString(body string) *Response {
	r.BodyBuffer = []byte(body)
	return r
}

// File defines the response body reading the data
// from disk based on the file path string.
func (r *Response) File(path string) *Response {
	r.BodyBuffer, r.Error = ioutil.ReadFile(path)
	return r
}

// JSON defines the response body based on a JSON based input.
func (r *Response) JSON(data interface{}) *Response {
	r.Header.Set("Content-Type", "application/json")
	r.BodyBuffer, r.Error = readAndDecode(data, "json")
	return r
}

// XML defines the response body based on a XML based input.
func (r *Response) XML(data interface{}) *Response {
	r.Header.Set("Content-Type", "application/xml")
	r.BodyBuffer, r.Error = readAndDecode(data, "xml")
	return r
}

// SetError defines the response simulated error.
func (r *Response) SetError(err error) *Response {
	r.Error = err
	return r
}

// Delay defines the response simulated delay.
// This feature is still experimental and will be improved in the future.
func (r *Response) Delay(delay time.Duration) *Response {
	r.ResponseDelay = delay
	return r
}

// Map adds a new response mapper function to map http.Response before the matching process.
func (r *Response) Map(fn MapResponseFunc) *Response {
	r.Mappers = append(r.Mappers, fn)
	return r
}

// Filter filters a new request filter function to filter http.Request before the matching process.
func (r *Response) Filter(fn FilterResponseFunc) *Response {
	r.Filters = append(r.Filters, fn)
	return r
}

// EnableNetworking enables the use real networking for the current mock.
func (r *Response) EnableNetworking() *Response {
	r.UseNetwork = true
	return r
}

// Done returns true if the mock was done and disabled.
func (r *Response) Done() bool {
	return r.Mock.Done()
}

func readAndDecode(data interface{}, kind string) ([]byte, error) {
	buf := &bytes.Buffer{}

	switch data.(type) {
	case string:
		buf.WriteString(data.(string))
	case []byte:
		buf.Write(data.([]byte))
	default:
		var err error
		if kind == "xml" {
			err = xml.NewEncoder(buf).Encode(data)
		} else {
			err = json.NewEncoder(buf).Encode(data)
		}
		if err != nil {
			return nil, err
		}
	}

	return ioutil.ReadAll(buf)
}
