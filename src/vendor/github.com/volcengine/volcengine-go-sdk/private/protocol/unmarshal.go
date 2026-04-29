package protocol

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"io"
	"io/ioutil"

	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

// UnmarshalDiscardBodyHandler is a named request handler to empty and close a response's volcenginebody
var UnmarshalDiscardBodyHandler = request.NamedHandler{Name: "volcenginesdk.shared.UnmarshalDiscardBody", Fn: UnmarshalDiscardBody}

// UnmarshalDiscardBody is a request handler to empty a response's volcenginebody and closing it.
func UnmarshalDiscardBody(r *request.Request) {
	if r.HTTPResponse == nil || r.HTTPResponse.Body == nil {
		return
	}

	io.Copy(ioutil.Discard, r.HTTPResponse.Body)
	r.HTTPResponse.Body.Close()
}
