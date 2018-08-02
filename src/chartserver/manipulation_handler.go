package chartserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	hlog "github.com/vmware/harbor/src/common/utils/log"
	helm_repo "k8s.io/helm/pkg/repo"
)

const (
	//NamespaceContextKey is context key for the namespace
	NamespaceContextKey ContextKey = ":repo"
)

//ContextKey is defined for add value in the context of http request
type ContextKey string

//ManipulationHandler includes all the handler methods for the purpose of manipulating the
//chart repository
type ManipulationHandler struct {
	//Proxy used to to transfer the traffic of requests
	//It's mainly used to talk to the backend chart server
	trafficProxy *ProxyEngine

	//Parse and process the chart version to provide required info data
	chartOperator *ChartOperator

	//HTTP client used to call the realted APIs of the backend chart repositories
	apiClient *ChartClient

	//Point to the url of the backend server
	backendServerAddress *url.URL

	//Cache the chart data
	chartCache *ChartCache
}

//ListCharts lists all the charts under the specified namespace
func (mh *ManipulationHandler) ListCharts(w http.ResponseWriter, req *http.Request) {
	url := strings.TrimPrefix(req.URL.String(), "/")
	url = fmt.Sprintf("%s/%s", mh.backendServerAddress.String(), url)

	content, err := mh.apiClient.GetContent(url)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	chartList, err := mh.chartOperator.GetChartList(content)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	jsonData, err := json.Marshal(chartList)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	writeJSONData(w, jsonData)
}

//GetChart returns all the chart versions under the specified chart
func (mh *ManipulationHandler) GetChart(w http.ResponseWriter, req *http.Request) {
	mh.trafficProxy.ServeHTTP(w, req)
}

//GetChartVersion get the specified version for one chart
//This handler should return the details of the chart version,
//maybe including metadata,dependencies and values etc.
func (mh *ManipulationHandler) GetChartVersion(w http.ResponseWriter, req *http.Request) {
	chartV, err := mh.getChartVersion(req.URL.String())
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	//Get and check namespace
	//even we get the data from cache
	var namespace string

	repoValue := req.Context().Value(NamespaceContextKey)
	if repoValue != nil {
		if ns, ok := repoValue.(string); ok {
			namespace = ns
		}
	}

	if len(strings.TrimSpace(namespace)) == 0 {
		WriteInternalError(w, errors.New("failed to extract namespace from the request"))
		return
	}

	//Query cache
	chartDetails := mh.chartCache.GetChart(chartV.Digest)
	if chartDetails == nil {
		//NOT hit!!
		content, err := mh.getChartVersionContent(namespace, chartV.URLs[0])
		if err != nil {
			WriteInternalError(w, err)
			return
		}

		//Process bytes and get more details of chart version
		chartDetails, err = mh.chartOperator.GetChartDetails(content)
		if err != nil {
			WriteInternalError(w, err)
			return
		}
		chartDetails.Metadata = chartV

		//Put it into the cache for next access
		mh.chartCache.PutChart(chartDetails)
	} else {
		//Just logged
		hlog.Debugf("Get detailed data from cache for chart: %s:%s (%s)",
			chartDetails.Metadata.Name,
			chartDetails.Metadata.Version,
			chartDetails.Metadata.Digest)
	}
	//The change of prov file will not cause any influence to the digest of chart,
	//and then the digital signature status should be not cached
	//
	//Generate the security report
	//prov file share same endpoint with the chart version
	//Just add .prov suffix to the chart version to form the path of prov file
	//Anyway, there will be a report about the digital signature status
	chartDetails.Security = &SecurityReport{
		Signature: &DigitalSignature{
			Signed: false,
		},
	}
	//Try to get the prov file to confirm if it is exitsing
	provFilePath := fmt.Sprintf("%s.prov", chartV.URLs[0])
	provBytes, err := mh.getChartVersionContent(namespace, provFilePath)
	if err == nil && len(provBytes) > 0 {
		chartDetails.Security.Signature.Signed = true
		chartDetails.Security.Signature.Provenance = provFilePath
	} else {
		//Just log it
		hlog.Errorf("Failed to get prov file for chart %s with error: %s, got %d bytes", chartV.Name, err.Error(), len(provBytes))
	}

	bytes, err := json.Marshal(chartDetails)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	writeJSONData(w, bytes)
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

//Get the basic metadata of chart version
func (mh *ManipulationHandler) getChartVersion(subPath string) (*helm_repo.ChartVersion, error) {
	url := fmt.Sprintf("%s/%s", mh.backendServerAddress.String(), strings.TrimPrefix(subPath, "/"))

	content, err := mh.apiClient.GetContent(url)
	if err != nil {
		return nil, err
	}

	chartVersion := &helm_repo.ChartVersion{}
	if err := yaml.Unmarshal(content, chartVersion); err != nil {
		return nil, err
	}

	return chartVersion, nil
}

//Get the content bytes of the chart version
func (mh *ManipulationHandler) getChartVersionContent(namespace string, subPath string) ([]byte, error) {
	url := path.Join(namespace, subPath)
	url = fmt.Sprintf("%s/%s", mh.backendServerAddress.String(), url)

	return mh.apiClient.GetContent(url)
}
