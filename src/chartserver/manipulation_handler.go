package chartserver

import (
	"net/http"
)

//ManipulationHandler includes all the handler methods for the purpose of manipulating the
//chart repository
type ManipulationHandler struct {
	//Proxy used to to transfer the traffic of requests
	//It's mainly used to talk to the backend chart server
	trafficProxy *ProxyEngine

	//Parse and process the chart version to provide required info data
	chartOperator *ChartOperator
}

//ListCharts lists all the charts under the specified namespace
func (mh *ManipulationHandler) ListCharts(w http.ResponseWriter, req *http.Request) {
	mh.trafficProxy.ServeHTTP(w, req)
}

//GetChart returns all the chart versions under the specified chart
func (mh *ManipulationHandler) GetChart(w http.ResponseWriter, req *http.Request) {
	mh.trafficProxy.ServeHTTP(w, req)
}

//GetChartVersion get the specified version for one chart
//This handler should return the details of the chart version,
//maybe including metadata,dependencies and values etc.
func (mh *ManipulationHandler) GetChartVersion(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("not implemented"))
}

//UploadChartVersion will save the new version of the chart to the backend storage
func (mh *ManipulationHandler) UploadChartVersion(w http.ResponseWriter, req *http.Request) {
	mh.trafficProxy.ServeHTTP(w, req)
}

//UploadProvenanceFile will save the provenance file of the chart to the backend storage
func (mh *ManipulationHandler) UploadProvenanceFile(w http.ResponseWriter, req *http.Request) {
	mh.trafficProxy.ServeHTTP(w, req)
}

//DeleteChartVersion will delete the specified version of the chart
func (mh *ManipulationHandler) DeleteChartVersion(w http.ResponseWriter, req *http.Request) {
	mh.trafficProxy.ServeHTTP(w, req)
}
