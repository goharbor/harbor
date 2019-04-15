package huawei

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/replication/model"
)

// FetchImages gets resources from Huawei SWR
func (adapter *Adapter) FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {

	resources := []*model.Resource{}

	urls := fmt.Sprintf("%s/dockyard/v2/repositories?filter=center::self", adapter.Registry.URL)

	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return resources, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")
	encodeAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", adapter.Registry.Credential.AccessKey, adapter.Registry.Credential.AccessSecret)))
	r.Header.Add("Authorization", "Basic "+encodeAuth)

	client := &http.Client{}
	if adapter.Registry.Insecure == true {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}
	resp, err := client.Do(r)
	if err != nil {
		return resources, err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resources, fmt.Errorf("[%d][%s]", code, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resources, err
	}
	repos := []hwRepoQueryResult{}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return resources, err
	}
	for _, repo := range repos {
		for _, namespace := range namespaces {
			if repo.NamespaceName == namespace {
				resource := parseRepoQueryResultToResource(repo)
				resource.Registry = adapter.Registry
				resources = append(resources, resource)
			}
		}
	}
	return resources, nil

}

func parseRepoQueryResultToResource(repo hwRepoQueryResult) *model.Resource {
	var resource model.Resource
	info := make(map[string]interface{})
	info["category"] = repo.Category
	info["description"] = repo.Description
	info["size"] = repo.Size
	info["is_public"] = repo.IsPublic
	info["num_images"] = repo.NumImages
	info["num_download"] = repo.NumDownload
	info["created_at"] = repo.CreatedAt
	info["updated_at"] = repo.UpdatedAt
	info["domain_name"] = repo.DomainName
	info["status"] = repo.Status
	info["total_range"] = repo.TotalRange

	repository := &model.Repository{
		Name:     repo.Name,
		Metadata: info,
	}
	resource.ExtendedInfo = info
	resource.Metadata = &model.ResourceMetadata{
		Repository: repository,
		Vtags:      repo.Tags,
		Labels:     []string{},
	}
	resource.Deleted = false
	resource.Override = false
	resource.Type = model.ResourceTypeRepository

	return &resource
}

type hwRepoQueryResult struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`

	Size         int64     `json:"size" `
	IsPublic     bool      `json:"is_public"`
	NumImages    int64     `json:"num_images"`
	NumDownload  int64     `json:"num_download"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Logo         string    `json:"logo"`
	LogoURL      string    `json:"url"`
	Path         string    `json:"path"`
	InternalPath string    `json:"internal_path"`

	DomainName    string   `json:"domain_name"`
	NamespaceName string   `json:"namespace"`
	Tags          []string `json:"tags"`
	Status        bool     `json:"status"`
	TotalRange    int64    `json:"total_range"`
}
