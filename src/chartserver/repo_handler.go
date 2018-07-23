package chartserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ghodss/yaml"
	"github.com/vmware/harbor/src/ui/filter"
	helm_repo "k8s.io/helm/pkg/repo"

	hlog "github.com/vmware/harbor/src/common/utils/log"
)

const (
	maxWorkers = 10
)

//RepositoryHandler defines all the handlers to handle the requests related with chart repository
//e.g: index.yaml and downloading chart objects
type RepositoryHandler struct {
	//Proxy used to to transfer the traffic of requests
	//It's mainly used to talk to the backend chart server
	trafficProxy *ProxyEngine

	//HTTP client used to call the realted APIs of the backend chart repositories
	apiClient *ChartClient

	//Point to the url of the backend server
	backendServerAddress *url.URL
}

//Pass work to the workers
//'index' is the location of processing namespace/project in the list
type workload struct {
	index uint32
}

//Result returned by worker
type processedResult struct {
	namespace       string
	indexFileOfRepo *helm_repo.IndexFile
}

//GetIndexFileWithNS will read the index.yaml data under the specified namespace
func (rh *RepositoryHandler) GetIndexFileWithNS(w http.ResponseWriter, req *http.Request) {
	rh.trafficProxy.ServeHTTP(w, req)
}

//GetIndexFile will read the index.yaml under all namespaces and merge them as a single one
//Please be aware that, to support this function, the backend chart repository server should
//enable multi-tenancies
func (rh *RepositoryHandler) GetIndexFile(w http.ResponseWriter, req *http.Request) {
	//Get project manager references
	projectMgr, err := filter.GetProjectManager(req)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	//Get all the projects
	results, err := projectMgr.List(nil)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	//If no projects existing, return empty index.yaml content immediately
	if results.Total == 0 {
		w.Write(emptyIndexFile())
		return
	}

	//The final merged index file
	mergedIndexFile := &helm_repo.IndexFile{
		APIVersion: "v1",
		Entries:    make(map[string]helm_repo.ChartVersions),
		Generated:  time.Now(),
		PublicKeys: []string{},
	}

	//Retrieve index.yaml for repositories
	workerPool := make(chan *workload, maxWorkers)
	//Sync the output results from the retriever
	resultChan := make(chan *processedResult, 1)
	//Receive error
	errorChan := make(chan error, 1)
	//Signal chan for merging work
	mergeDone := make(chan struct{}, 1)
	//Total projects/namespaces
	total := uint32(results.Total)
	//Track all the background threads
	waitGroup := new(sync.WaitGroup)

	//Initialize
	initialItemCount := maxWorkers
	if total < maxWorkers {
		initialItemCount = int(total)
	}
	for i := 0; i < initialItemCount; i++ {
		workerPool <- &workload{uint32(i)}
	}

	//Atomtic index
	var indexRef uint32
	atomic.AddUint32(&indexRef, uint32(initialItemCount-1))

	//Start the index files merging thread
	go func() {
		defer func() {
			mergeDone <- struct{}{}
		}()

		for res := range resultChan {
			rh.mergeIndexFile(res.namespace, mergedIndexFile, res.indexFileOfRepo)
		}
	}()

	//Retrieve the index files for the repositories
	//and blocking here
LOOP:
	for {
		select {
		case work := <-workerPool:
			if work.index >= total {
				break LOOP
			}
			//Process
			//New one
			waitGroup.Add(1)
			namespace := results.Projects[work.index].Name
			go func(ns string) {
				//Return the worker back to the pool
				defer func() {
					waitGroup.Done() //done

					//Put one. The invalid index will be treated as a signal to quit loop
					nextOne := atomic.AddUint32(&indexRef, 1)
					workerPool <- &workload{nextOne}
				}()

				indexFile, err := rh.getIndexYamlWithNS(ns)
				if err != nil {
					errorChan <- err
					return
				}

				//Output
				resultChan <- &processedResult{
					namespace:       ns,
					indexFileOfRepo: indexFile,
				}
			}(namespace)
		case err = <-errorChan:
			//Quit earlier
			break LOOP
		case <-req.Context().Done():
			//Quit earlier
			err = errors.New("request of getting index yaml file is aborted")
			break LOOP
		}
	}

	//Hold util all the retrieving work are done
	waitGroup.Wait()

	//close consumer channel
	close(resultChan)

	//Wait until merging thread quit
	<-mergeDone

	//All the threads are done
	//Met an error
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	//Remove duplicated keys in public key list
	hash := make(map[string]string)
	for _, key := range mergedIndexFile.PublicKeys {
		hash[key] = key
	}
	mergedIndexFile.PublicKeys = []string{}
	for k := range hash {
		mergedIndexFile.PublicKeys = append(mergedIndexFile.PublicKeys, k)
	}

	bytes, err := yaml.Marshal(mergedIndexFile)
	if err != nil {
		WriteInternalError(w, err)
		return
	}

	w.Write(bytes)
}

//DownloadChartObject will download the stored chart object to the client
//e.g: helm install
func (rh *RepositoryHandler) DownloadChartObject(w http.ResponseWriter, req *http.Request) {
	rh.trafficProxy.ServeHTTP(w, req)
}

//Get the index yaml file under the specified namespace from the backend server
func (rh *RepositoryHandler) getIndexYamlWithNS(namespace string) (*helm_repo.IndexFile, error) {
	//Join url path
	url := path.Join(namespace, "index.yaml")
	url = fmt.Sprintf("%s/%s", rh.backendServerAddress.String(), url)
	hlog.Debugf("Getting index.yaml from '%s'", url)

	content, err := rh.apiClient.GetContent(url)
	if err != nil {
		return nil, err
	}

	//Traverse to index file object for merging
	indexFile := helm_repo.NewIndexFile()
	if err := yaml.Unmarshal(content, indexFile); err != nil {
		return nil, err
	}

	return indexFile, nil
}

//Merge the content of mergingIndexFile to the baseIndex
//The chart url should be without --chart-url prefix
func (rh *RepositoryHandler) mergeIndexFile(namespace string,
	baseIndex *helm_repo.IndexFile,
	mergingIndexFile *helm_repo.IndexFile) {
	//Append entries
	for chartName, chartVersions := range mergingIndexFile.Entries {
		nameWithNS := fmt.Sprintf("%s/%s", namespace, chartName)
		for _, version := range chartVersions {
			version.Name = nameWithNS
			//Currently there is only one url
			for index, url := range version.URLs {
				version.URLs[index] = path.Join(namespace, url)
			}
		}

		//Appended
		baseIndex.Entries[nameWithNS] = chartVersions
	}

	//Update generated time
	if mergingIndexFile.Generated.After(baseIndex.Generated) {
		baseIndex.Generated = mergingIndexFile.Generated
	}

	//Merge public keys
	baseIndex.PublicKeys = append(baseIndex.PublicKeys, mergingIndexFile.PublicKeys...)
}

//Generate empty index file
func emptyIndexFile() []byte {
	emptyIndexFile := &helm_repo.IndexFile{}
	emptyIndexFile.Generated = time.Now()

	//Ignore the error
	rawData, _ := json.Marshal(emptyIndexFile)

	return rawData
}
