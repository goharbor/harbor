package chartserver

import (
	"net/http"
	"net/http/httputil"
)

//RepositoryHandler defines all the handlers to handle the requests related with chart repository
//e.g: index.yaml and downloading chart objects
type RepositoryHandler struct {
	//Proxy used to to transfer the traffic of requests
	//It's mainly used to talk to the backend chart server
	trafficProxy *httputil.ReverseProxy
}

//GetIndexFileWithNS will read the index.yaml data under the specified namespace
func (rh *RepositoryHandler) GetIndexFileWithNS(w http.ResponseWriter, req *http.Request) {
}

//GetIndexFile will read the index.yaml under all namespaces and merge them as a single one
//Please be aware that, to support this function, the backend chart repository server should
//enable multi-tenancies
func (rh *RepositoryHandler) GetIndexFile(w http.ResponseWriter, req *http.Request) {}

//DownloadChartObject will download the stored chart object to the client
//e.g: helm install
func (rh *RepositoryHandler) DownloadChartObject(w http.ResponseWriter, req *http.Request) {}
