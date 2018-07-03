package chartserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	userName            = "chart_controller"
	passwordKey         = "UI_SECRET"
	agentHarbor         = "HARBOR"
	authHeader          = "Authorization"
	contentLengthHeader = "Content-Length"
)

//ProxyEngine is used to proxy the related traffics
type ProxyEngine struct {
	//The backend target server the traffic will be forwarded to
	//Just in case we'll use it
	backend *url.URL

	//Use go reverse proxy as engine
	engine *httputil.ReverseProxy
}

//NewProxyEngine is constructor of NewProxyEngine
func NewProxyEngine(target *url.URL) *ProxyEngine {
	return &ProxyEngine{
		backend: target,
		engine: &httputil.ReverseProxy{
			ErrorLog: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
			Director: func(req *http.Request) {
				director(target, req)
			},
			ModifyResponse: modifyResponse,
		},
	}
}

//ServeHTTP serves the incoming http requests
func (pe *ProxyEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pe.engine.ServeHTTP(w, req)
}

//Overwrite the http requests
func director(target *url.URL, req *http.Request) {
	targetQuery := target.RawQuery

	//Overwrite the request URL to the target path
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", agentHarbor)
	}

	//Get the password from the env
	//Ignore the empty checking, the backend server should return the right status code
	//with invalid credential
	password := os.Getenv(passwordKey)

	//Add authentication header
	req.SetBasicAuth(userName, password)
}

//Modify the http response
func modifyResponse(res *http.Response) error {
	//Detect the 401 code, if it is,
	//overwrite it to 500.
	//We also re-write the error content
	if res.StatusCode == http.StatusUnauthorized {
		errorObj := make(map[string]string)
		errorObj["error"] = "operation request from unauthentic source is rejected"
		content, err := json.Marshal(errorObj)
		if err != nil {
			return err
		}

		size := len(content)
		body := ioutil.NopCloser(bytes.NewReader(content))
		res.Body = body
		res.ContentLength = int64(size)
		res.Header.Set(contentLengthHeader, strconv.Itoa(size))
		res.StatusCode = http.StatusInternalServerError
	}

	return nil
}

//Join the path
//Copy from the go reverse proxy
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
