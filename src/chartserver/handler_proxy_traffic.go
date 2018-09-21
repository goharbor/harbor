package chartserver

import (
	"net/http"
)

// ProxyTraffic implements the interface method.
func (c *Controller) ProxyTraffic(w http.ResponseWriter, req *http.Request) {
	if c.trafficProxy != nil {
		c.trafficProxy.ServeHTTP(w, req)
	}
}
