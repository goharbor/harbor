package client

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http/httputil"
	"strings"

	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const logReqMsg = `DEBUG: Request %s/%s Details:
---[ REQUEST POST-SIGN ]-----------------------------
%s
-----------------------------------------------------`

const logReqErrMsg = `DEBUG ERROR: Request %s/%s:
---[ REQUEST DUMP ERROR ]-----------------------------
%s
------------------------------------------------------`

type logWriter struct {
	// Logger is what we will use to log the payload of a response.
	Logger volcengine.Logger
	// buf stores the contents of what has been read
	buf *bytes.Buffer
}

func (logger *logWriter) Write(b []byte) (int, error) {
	return logger.buf.Write(b)
}

type teeReaderCloser struct {
	// io.Reader will be a tee reader that is used during logging.
	// This structure will read from a volcenginebody and write the contents to a logger.
	io.Reader
	// Source is used just to close when we are done reading.
	Source io.ReadCloser
}

func (reader *teeReaderCloser) Close() error {
	return reader.Source.Close()
}

type LogStruct struct {
	Level         string
	OperationName string
	Request       interface{} `json:"Request,omitempty"`
	Body          interface{} `json:"Body,omitempty"`
	Response      interface{} `json:"Response,omitempty"`
	Type          string
	AccountId     string `json:"AccountId,omitempty"`
	Context       context.Context
}

func logStructLog(r *request.Request, level string, logStruct LogStruct) {
	logStruct.Level = level
	if r.IsJsonBody && strings.HasSuffix(logStruct.OperationName, "Request") {
		logStruct.Body = r.Params
	}
	if r.Config.LogAccount != nil {
		logStruct.AccountId = *r.Config.LogAccount(r.Context())
	}
	//b, _ := json.Marshal(logStruct)
	r.Config.Logger.Log(logStruct)
}

var LogInputHandler = request.NamedHandler{
	Name: "volcenginesdk.client.LogInput",
	Fn:   logInput,
}

func logInput(r *request.Request) {
	logInfoStruct := r.Config.LogLevel.Matches(volcengine.LogInfoWithInputAndOutput)
	logDebugStruct := r.Config.LogLevel.Matches(volcengine.LogDebugWithInputAndOutput)

	logStruct := LogStruct{
		OperationName: r.Operation.Name,
		Type:          "Request",
		Request:       r.Input,
		Context:       r.Context(),
	}

	if logInfoStruct {
		logStructLog(r, "INFO", logStruct)
	} else if logDebugStruct {
		logStructLog(r, "DEBUG", logStruct)
	}
}

var LogOutHandler = request.NamedHandler{
	Name: "volcenginesdk.client.LogOutput",
	Fn:   LogOutput,
}

func LogOutput(r *request.Request) {
	logInfoStruct := r.Config.LogLevel.Matches(volcengine.LogInfoWithInputAndOutput)
	logDebugStruct := r.Config.LogLevel.Matches(volcengine.LogDebugWithInputAndOutput)

	logStruct := LogStruct{
		OperationName: r.Operation.Name,
		Response:      r.Data,
		Type:          "Response",
		Context:       r.Context(),
	}

	if logInfoStruct {
		logStructLog(r, "INFO", logStruct)
	} else if logDebugStruct {
		logStructLog(r, "DEBUG", logStruct)
	}
}

// LogHTTPRequestHandler is a SDK request handler to log the HTTP request sent
// to a service. Will include the HTTP request volcenginebody if the LogLevel of the
// request matches LogDebugWithHTTPBody.
var LogHTTPRequestHandler = request.NamedHandler{
	Name: "volcenginesdk.client.LogRequest",
	Fn:   logRequest,
}

func logRequest(r *request.Request) {
	logBody := r.Config.LogLevel.Matches(volcengine.LogDebugWithHTTPBody)
	bodySeekable := volcengine.IsReaderSeekable(r.Body)

	b, err := httputil.DumpRequestOut(r.HTTPRequest, logBody)
	if err != nil {
		r.Config.Logger.Log(fmt.Sprintf(logReqErrMsg,
			r.ClientInfo.ServiceName, r.Operation.Name, err))
		return
	}

	if logBody {
		if !bodySeekable {
			r.SetReaderBody(volcengine.ReadSeekCloser(r.HTTPRequest.Body))
		}
		// Reset the request volcenginebody because dumpRequest will re-wrap the
		// r.HTTPRequest's Body as a NoOpCloser and will not be reset after
		// read by the HTTP client reader.
		if err := r.Error; err != nil {
			r.Config.Logger.Log(fmt.Sprintf(logReqErrMsg,
				r.ClientInfo.ServiceName, r.Operation.Name, err))
			return
		}
	}

	r.Config.Logger.Log(fmt.Sprintf(logReqMsg,
		r.ClientInfo.ServiceName, r.Operation.Name, string(b)))
}

// LogHTTPRequestHeaderHandler is a SDK request handler to log the HTTP request sent
// to a service. Will only log the HTTP request's headers. The request payload
// will not be read.
var LogHTTPRequestHeaderHandler = request.NamedHandler{
	Name: "volcenginesdk.client.LogRequestHeader",
	Fn:   logRequestHeader,
}

func logRequestHeader(r *request.Request) {
	b, err := httputil.DumpRequestOut(r.HTTPRequest, false)
	if err != nil {
		r.Config.Logger.Log(fmt.Sprintf(logReqErrMsg,
			r.ClientInfo.ServiceName, r.Operation.Name, err))
		return
	}

	r.Config.Logger.Log(fmt.Sprintf(logReqMsg,
		r.ClientInfo.ServiceName, r.Operation.Name, string(b)))
}

const logRespMsg = `DEBUG: Response %s/%s Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`

const logRespErrMsg = `DEBUG ERROR: Response %s/%s:
---[ RESPONSE DUMP ERROR ]-----------------------------
%s
-----------------------------------------------------`

// LogHTTPResponseHandler is a SDK request handler to log the HTTP response
// received from a service. Will include the HTTP response volcenginebody if the LogLevel
// of the request matches LogDebugWithHTTPBody.
var LogHTTPResponseHandler = request.NamedHandler{
	Name: "volcenginesdk.client.LogResponse",
	Fn:   logResponse,
}

func logResponse(r *request.Request) {
	lw := &logWriter{r.Config.Logger, bytes.NewBuffer(nil)}

	if r.HTTPResponse == nil {
		lw.Logger.Log(fmt.Sprintf(logRespErrMsg,
			r.ClientInfo.ServiceName, r.Operation.Name, "request's HTTPResponse is nil"))
		return
	}

	logBody := r.Config.LogLevel.Matches(volcengine.LogDebugWithHTTPBody)
	if logBody {
		r.HTTPResponse.Body = &teeReaderCloser{
			Reader: io.TeeReader(r.HTTPResponse.Body, lw),
			Source: r.HTTPResponse.Body,
		}
	}

	handlerFn := func(req *request.Request) {
		b, err := httputil.DumpResponse(req.HTTPResponse, false)
		if err != nil {
			lw.Logger.Log(fmt.Sprintf(logRespErrMsg,
				req.ClientInfo.ServiceName, req.Operation.Name, err))
			return
		}

		lw.Logger.Log(fmt.Sprintf(logRespMsg,
			req.ClientInfo.ServiceName, req.Operation.Name, string(b)))

		if logBody {
			b, err := ioutil.ReadAll(lw.buf)
			if err != nil {
				lw.Logger.Log(fmt.Sprintf(logRespErrMsg,
					req.ClientInfo.ServiceName, req.Operation.Name, err))
				return
			}

			lw.Logger.Log(string(b))
		}
	}

	const handlerName = "volcenginesdk.client.LogResponse.ResponseBody"

	r.Handlers.Unmarshal.SetBackNamed(request.NamedHandler{
		Name: handlerName, Fn: handlerFn,
	})
	r.Handlers.UnmarshalError.SetBackNamed(request.NamedHandler{
		Name: handlerName, Fn: handlerFn,
	})
}

// LogHTTPResponseHeaderHandler is a SDK request handler to log the HTTP
// response received from a service. Will only log the HTTP response's headers.
// The response payload will not be read.
var LogHTTPResponseHeaderHandler = request.NamedHandler{
	Name: "volcenginesdk.client.LogResponseHeader",
	Fn:   logResponseHeader,
}

func logResponseHeader(r *request.Request) {
	if r.Config.Logger == nil {
		return
	}

	b, err := httputil.DumpResponse(r.HTTPResponse, false)
	if err != nil {
		r.Config.Logger.Log(fmt.Sprintf(logRespErrMsg,
			r.ClientInfo.ServiceName, r.Operation.Name, err))
		return
	}

	r.Config.Logger.Log(fmt.Sprintf(logRespMsg,
		r.ClientInfo.ServiceName, r.Operation.Name, string(b)))
}
