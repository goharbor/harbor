package chartserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/ghodss/yaml"
	"github.com/goharbor/harbor/src/ui/filter"
	"k8s.io/helm/cmd/helm/search"
	helm_repo "k8s.io/helm/pkg/repo"

	hlog "github.com/goharbor/harbor/src/common/utils/log"
)

const (
	maxWorkers = 10

	//Keep consistent with 'helm search' command
	searchMaxScore = 25
)

// RepositoryHandler defines all the handlers to handle the requests related with chart repository
// e.g: index.yaml and downloading chart objects
type RepositoryHandler struct {
	// Proxy used to to transfer the traffic of requests
	// It's mainly used to talk to the backend chart server
	trafficProxy *ProxyEngine

	// HTTP client used to call the realted APIs of the backend chart repositories
	apiClient *ChartClient

	// Point to the url of the backend server
	backendServerAddress *url.URL
}

// Result returned by worker
type processedResult struct {
	namespace       string
	indexFileOfRepo *helm_repo.IndexFile
}

// GetIndexFileWithNS will read the index.yaml data under the specified namespace
func (rh *RepositoryHandler) GetIndexFileWithNS(w http.ResponseWriter, req *http.Request) {
	rh.trafficProxy.ServeHTTP(w, req)
}

// GetIndexFile will read the index.yaml under all namespaces and merge them as a single one
// Please be aware that, to support this function, the backend chart repository server should
// enable multi-tenancies
func (rh *RepositoryHandler) GetIndexFile(w http.ResponseWriter, req *http.Request) {
	// Get project manager references
	projectMgr, err := filter.GetProjectManager(req)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	// Get all the projects
	results, err := projectMgr.List(nil)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	// If no projects existing, return empty index.yaml content immediately
	if results.Total == 0 {
		w.Write(emptyIndexFile())
		return
	}

	namespaces := []string{}
	for _, p := range results.Projects {
		namespaces = append(namespaces, p.Name)
	}

	mergedIndexFile, err := rh.getIndexYaml(namespaces)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	bytes, err := yaml.Marshal(mergedIndexFile)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	w.Write(bytes)
}

// DownloadChartObject will download the stored chart object to the client
// e.g: helm install
func (rh *RepositoryHandler) DownloadChartObject(w http.ResponseWriter, req *http.Request) {
	rh.trafficProxy.ServeHTTP(w, req)
}

// SearchChart search charts in the specified namespaces with the keyword q.
// RegExp mode is enabled as default.
// For each chart, only the latest version will shown in the result list if matched to avoid duplicated entries.
// Keep consistent with `helm search` command.
func (rh *RepositoryHandler) SearchChart(q string, namespaces []string) ([]*search.Result, error) {
	if len(q) == 0 || len(namespaces) == 0 {
		// Return empty list
		return []*search.Result{}, nil
	}

	// Get the merged index yaml file of the namespaces
	ind, err := rh.getIndexYaml(namespaces)
	if err != nil {
		return nil, err
	}

	// Build the search index
	index := search.NewIndex()
	// As the repo name is already merged into the index yaml, we use empty repo name.
	// Set 'All' to false to return only one version for each chart.
	index.AddRepo("", ind, false)

	// Search
	// RegExp is enabled
	results, err := index.Search(q, searchMaxScore, true)
	if err != nil {
		return nil, err
	}

	// Sort results.
	search.SortScore(results)

	return results, nil
}

// getIndexYaml will get the index yaml files for all the namespaces and merge them
// as one unified index yaml file.
func (rh *RepositoryHandler) getIndexYaml(namespaces []string) (*helm_repo.IndexFile, error) {
	// The final merged index file
	mergedIndexFile := &helm_repo.IndexFile{
		APIVersion: "v1",
		Entries:    make(map[string]helm_repo.ChartVersions),
		Generated:  time.Now(),
		PublicKeys: []string{},
	}

	// Sync the output results from the retriever
	resultChan := make(chan *processedResult, 1)
	// Receive error
	errorChan := make(chan error, 1)
	// Signal chan for merging work
	mergeDone := make(chan struct{}, 1)
	// Total projects/namespaces
	total := len(namespaces)
	// Initialize
	initialItemCount := maxWorkers
	if total < maxWorkers {
		initialItemCount = total
	}
	// Retrieve index.yaml for repositories
	workerPool := make(chan struct{}, initialItemCount)

	// Add initial tokens to the pool
	for i := 0; i < initialItemCount; i++ {
		workerPool <- struct{}{}
	}
	// Track all the background threads
	waitGroup := new(sync.WaitGroup)

	// Start the index files merging thread
	go func() {
		defer func() {
			mergeDone <- struct{}{}
		}()

		for res := range resultChan {
			rh.mergeIndexFile(res.namespace, mergedIndexFile, res.indexFileOfRepo)
		}
	}()

	// Retrieve the index files for the repositories
	// and blocking here
	var err error
LOOP:
	for _, ns := range namespaces {
		// Check if error has occurred in some goroutines
		select {
		case err = <-errorChan:
			break LOOP
		default:
			// do nothing
		}

		// Apply one token before processing
		<-workerPool

		waitGroup.Add(1)
		go func(ns string) {
			defer func() {
				waitGroup.Done() //done
				//Return the worker back to the pool
				workerPool <- struct{}{}
			}()

			indexFile, err := rh.getIndexYamlWithNS(ns)
			if err != nil {
				if len(errorChan) == 0 {
					//Only need one error as failure signal
					errorChan <- err
				}
				return
			}

			// Output
			resultChan <- &processedResult{
				namespace:       ns,
				indexFileOfRepo: indexFile,
			}
		}(ns)
	}

	// Hold util all the retrieving work are done
	waitGroup.Wait()

	// close merge channel
	close(resultChan)

	// Wait until merging thread quit
	<-mergeDone

	// All the threads are done
	// Make sure error in the chan is read
	if err == nil && len(errorChan) > 0 {
		err = <-errorChan
	}

	// Met an error
	if err != nil {
		return nil, err
	}

	// Remove duplicated keys in public key list
	hash := make(map[string]string)
	for _, key := range mergedIndexFile.PublicKeys {
		hash[key] = key
	}
	mergedIndexFile.PublicKeys = []string{}
	for k := range hash {
		mergedIndexFile.PublicKeys = append(mergedIndexFile.PublicKeys, k)
	}

	return mergedIndexFile, nil
}

// Get the index yaml file under the specified namespace from the backend server
func (rh *RepositoryHandler) getIndexYamlWithNS(namespace string) (*helm_repo.IndexFile, error) {
	// Join url path
	url := path.Join(namespace, "index.yaml")
	url = fmt.Sprintf("%s/%s", rh.backendServerAddress.String(), url)
	hlog.Debugf("Getting index.yaml from '%s'", url)

	content, err := rh.apiClient.GetContent(url)
	if err != nil {
		return nil, err
	}

	// Traverse to index file object for merging
	indexFile := helm_repo.NewIndexFile()
	if err := yaml.Unmarshal(content, indexFile); err != nil {
		return nil, err
	}

	return indexFile, nil
}

// Merge the content of mergingIndexFile to the baseIndex
// The chart url should be without --chart-url prefix
func (rh *RepositoryHandler) mergeIndexFile(namespace string,
	baseIndex *helm_repo.IndexFile,
	mergingIndexFile *helm_repo.IndexFile) {
	// Append entries
	for chartName, chartVersions := range mergingIndexFile.Entries {
		nameWithNS := fmt.Sprintf("%s/%s", namespace, chartName)
		for _, version := range chartVersions {
			version.Name = nameWithNS
			// Currently there is only one url
			for index, url := range version.URLs {
				version.URLs[index] = path.Join(namespace, url)
			}
		}

		// Appended
		baseIndex.Entries[nameWithNS] = chartVersions
	}

	// Update generated time
	if mergingIndexFile.Generated.After(baseIndex.Generated) {
		baseIndex.Generated = mergingIndexFile.Generated
	}

	// Merge public keys
	baseIndex.PublicKeys = append(baseIndex.PublicKeys, mergingIndexFile.PublicKeys...)
}

// Generate empty index file
func emptyIndexFile() []byte {
	emptyIndexFile := &helm_repo.IndexFile{}
	emptyIndexFile.Generated = time.Now()

	// Ignore the error
	rawData, _ := json.Marshal(emptyIndexFile)

	return rawData
}
