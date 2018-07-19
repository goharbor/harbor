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
)

const (
	agentHarbor         = "HARBOR"
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
func NewProxyEngine(target *url.URL, cred *Credential) *ProxyEngine {
	return &ProxyEngine{
		backend: target,
		engine: &httputil.ReverseProxy{
			ErrorLog: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
			Director: func(req *http.Request) {
				director(target, cred, req)
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
func director(target *url.URL, cred *Credential, req *http.Request) {
	//Closure
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

	//Add authentication header if it is existing
	if cred != nil {
		req.SetBasicAuth(cred.Username, cred.Password)
	}
}

//Modify the http response
func modifyResponse(res *http.Response) error {
	//Accept cases
	//Success or redirect
	if res.StatusCode >= http.StatusOK && res.StatusCode <= http.StatusTemporaryRedirect {
		return nil
	}

	//Detect the 401 code, if it is,overwrite it to 500.
	//We also re-write the error content to structural error object
	errorObj := make(map[string]string)
	if res.StatusCode == http.StatusUnauthorized {
		errorObj["error"] = "operation request from unauthorized source is rejected"
		res.StatusCode = http.StatusInternalServerError
	} else {
		//Extract the error and wrap it into the error object
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
