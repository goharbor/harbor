package chartserver

import (
	"net/http"
)

//BaseHandler defines the handlers related with the chart server itself.
type BaseHandler struct {
	//Proxy used to to transfer the traffic of requests
	//It's mainly used to talk to the backend chart server
	trafficProxy *ProxyEngine
}

//GetHealthStatus will return the health status of the backend chart repository server
func (bh *BaseHandler) GetHealthStatus(w http.ResponseWriter, req *http.Request) {
	bh.trafficProxy.ServeHTTP(w, req)
}
