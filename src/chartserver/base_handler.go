package chartserver

import (
	"net/http"
	"net/http/httputil"
)

//BaseHandler defines the handlers related with the chart server itself.
type BaseHandler struct {
	//Proxy used to to transfer the traffic of requests
	//It's mainly used to talk to the backend chart server
	trafficProxy *httputil.ReverseProxy
}

//GetHealthStatus will return the health status of the backend chart repository server
func (bh *BaseHandler) GetHealthStatus(w http.ResponseWriter, req *http.Request) {}
