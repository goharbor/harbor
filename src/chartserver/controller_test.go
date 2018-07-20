package chartserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	helm_repo "k8s.io/helm/pkg/repo"
)

//Prepare, start the mock servers
func TestStartServers(t *testing.T) {
	if err := startMockServers(); err != nil {
		t.Fatal(err)
	}
}

//Test /health
func TestGetHealthOfBaseHandler(t *testing.T) {
	content, err := httpClient.GetContent(fmt.Sprintf("%s/health", getTheAddrOfFrontServer()))
	if err != nil {
		t.Fatal(err)
	}

	status := make(map[string]interface{})
	if err := json.Unmarshal(content, &status); err != nil {
		t.Fatalf("Unmarshal error: %s, %s", err, content)
	}
	healthy, ok := status["health"].(bool)
	if !ok || !healthy {
		t.Fatalf("Expect healthy of server to be 'true' but got %v", status["health"])
	}
}

//Get /repo1/index.yaml
func TestGetIndexYamlByRepo(t *testing.T) {
	indexFile, err := getIndexYaml("/repo1/index.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(indexFile.Entries) != 3 {
		t.Fatalf("Expect index file with 3 entries, but got %d", len(indexFile.Entries))
	}
}

//Test get /index.yaml
func TestGetUnifiedYamlFile(t *testing.T) {
	indexFile, err := getIndexYaml("/index.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(indexFile.Entries) != 5 {
		t.Fatalf("Expect index file with 5 entries, but got %d", len(indexFile.Entries))
	}

	_, ok := indexFile.Entries["repo1/harbor"]
	if !ok {
		t.Fatal("Expect chart entry 'repo1/harbor' but got nothing")
	}

	_, ok = indexFile.Entries["repo2/harbor"]
	if !ok {
		t.Fatal("Expect chart entry 'repo2/harbor' but got nothing")
	}
}

//Test download /:repo/charts/chart.tar
//Use this case to test the proxy function
func TestDownloadChart(t *testing.T) {
	content, err := httpClient.GetContent(fmt.Sprintf("%s/repo1/charts/harbor-0.2.0.tgz", getTheAddrOfFrontServer()))
	if err != nil {
		t.Fatal(err)
	}

	gotSize := len(content)
	expectSize := len(helmChartContent)

	if gotSize != expectSize {
		t.Fatalf("Expect %d bytes data but got %d bytes", expectSize, gotSize)
	}
}

//Test get /api/:repo/charts
func TestRetrieveChartList(t *testing.T) {
	content, err := httpClient.GetContent(fmt.Sprintf("%s/api/repo1/charts", getTheAddrOfFrontServer()))
	if err != nil {
		t.Fatal(err)
	}

	chartList := make([]*ChartInfo, 0)
	err = json.Unmarshal(content, &chartList)
	if err != nil {
		t.Fatalf("Unmarshal error: %s", err)
	}

	if len(chartList) != 2 {
		t.Fatalf("Expect to get 2 charts in the list but got %d", len(chartList))
	}

	foundItem := false
	for _, chartInfo := range chartList {
		if chartInfo.Name == "hello-helm" && chartInfo.TotalVersions == 2 {
			foundItem = true
			break
		}
	}

	if !foundItem {
		t.Fatalf("Expect chart 'hello-helm' with 2 versions but got nothing")
	}
}

//Test get /api/:repo/charts/:chart_name/:version
func TestGetChartVersion(t *testing.T) {
	content, err := httpClient.GetContent(fmt.Sprintf("%s/api/repo1/charts/harbor/0.2.0", getTheAddrOfFrontServer()))
	if err != nil {
		t.Fatal(err)
	}

	chartVersion := &ChartVersionDetails{}
	if err = json.Unmarshal(content, chartVersion); err != nil {
		t.Fatalf("Unmarshal error: %s", err)
	}

	if chartVersion.Metadata.Name != "harbor" {
		t.Fatalf("Expect harbor chart version but got %s", chartVersion.Metadata.Name)
	}

	if chartVersion.Metadata.Version != "0.2.0" {
		t.Fatalf("Expect version '0.2.0' but got version %s", chartVersion.Metadata.Version)
	}

	if len(chartVersion.Dependencies) != 1 {
		t.Fatalf("Expect 1 dependency but got %d ones", len(chartVersion.Dependencies))
	}

	if len(chartVersion.Values) != 99 {
		t.Fatalf("Expect 99 k-v values but got %d", len(chartVersion.Values))
	}
}

//Test get /api/:repo/charts/:chart_name/:version with none-existing version
func TestGetChartVersionWithError(t *testing.T) {
	_, err := httpClient.GetContent(fmt.Sprintf("%s/api/repo1/charts/harbor/1.0.0", getTheAddrOfFrontServer()))
	if err == nil {
		t.Fatal("Expect an error but got nil")
	}
}

//Get /api/repo1/charts/harbor
//401 will be rewritten to 500 with specified error
func TestResponseRewrite(t *testing.T) {
	response, err := http.Get(fmt.Sprintf("%s/api/repo1/charts/harbor", getTheAddrOfFrontServer()))
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expect status code 500 but got %d", response.StatusCode)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Read bytes from http response failed with error: %s", err)
	}
	defer response.Body.Close()

	errObj := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &errObj); err != nil {
		t.Fatalf("Unmarshal error: %s", err)
	}

	if msg, ok := errObj["error"]; !ok {
		t.Fatal("Expect an error message from server but got nothing")
	} else {
		if !strings.Contains(msg.(string), "operation request from unauthorized source is rejected") {
			t.Fatal("Missing the required error message")
		}
	}
}

//Clear environments
func TestStopServers(t *testing.T) {
	stopMockServers()
}

//Utility method for getting index yaml file
func getIndexYaml(path string) (*helm_repo.IndexFile, error) {
	content, err := httpClient.GetContent(fmt.Sprintf("%s%s", getTheAddrOfFrontServer(), path))
	if err != nil {
		return nil, err
	}

	indexFile := &helm_repo.IndexFile{}
	if err := yaml.Unmarshal(content, indexFile); err != nil {
		return nil, fmt.Errorf("Unmarshal error: %s", err)
	}

	if indexFile == nil {
		return nil, fmt.Errorf("Got nil index yaml file")
	}

	return indexFile, nil
}
