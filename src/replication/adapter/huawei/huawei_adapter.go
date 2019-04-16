package huawei

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/goharbor/harbor/src/replication/util"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
)

const (
	huawei model.RegistryType = "Huawei"
)

func init() {
	err := adapter.RegisterFactory(huawei, AdapterFactory)
	if err != nil {
		log.Errorf("failed to register factory for Huawei: %v", err)
		return
	}
	log.Infof("the factory of Huawei adapter was registered")
}

// Adapter is for images replications between harbor and Huawei image repository(SWR)
type Adapter struct {
	Registry *model.Registry
}

// Info gets info about Huawei SWR
func (adapter Adapter) Info() (*model.RegistryInfo, error) {
	registryInfo := model.RegistryInfo{
		Type:                     huawei,
		Description:              "Adapter for SWR -- The image registry of Huawei Cloud",
		SupportedResourceTypes:   []model.ResourceType{model.ResourceTypeRepository},
		SupportedResourceFilters: []*model.FilterStyle{},
		SupportedTriggers:        []model.TriggerType{},
	}
	return &registryInfo, nil
}

// ListNamespaces lists namespaces from Huawei SWR with the provided query conditions.
func (adapter Adapter) ListNamespaces(query *model.NamespaceQuery) ([]*model.Namespace, error) {
	var namespaces []*model.Namespace

	urls := fmt.Sprintf("%s/dockyard/v2/visible/namespaces", adapter.Registry.URL)

	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return namespaces, err
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
		return namespaces, err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return namespaces, fmt.Errorf("[%d][%s]", code, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return namespaces, err
	}

	var namespacesData hwNamespaceList
	err = json.Unmarshal(body, &namespacesData)
	if err != nil {
		return namespaces, err
	}
	reg := fmt.Sprintf(".*%s.*", strings.Replace(query.Name, " ", "", -1))

	for _, namespaceData := range namespacesData.Namespace {
		namespace := model.Namespace{
			Name:     namespaceData.Name,
			Metadata: namespaceData.metadata(),
		}
		b, err := regexp.MatchString(reg, namespace.Name)
		if err != nil {
			return namespaces, nil
		}
		if b {
			namespaces = append(namespaces, &namespace)
		}
	}
	return namespaces, nil
}

// ConvertResourceMetadata convert resource metadata for Huawei SWR
func (adapter Adapter) ConvertResourceMetadata(resourceMetadata *model.ResourceMetadata, namespace *model.Namespace) (*model.ResourceMetadata, error) {
	metadata := &model.ResourceMetadata{
		Repository: resourceMetadata.Repository,
		Vtags:      resourceMetadata.Vtags,
		Labels:     resourceMetadata.Labels,
	}
	return metadata, nil
}

// PrepareForPush prepare for push to Huawei SWR
func (adapter Adapter) PrepareForPush(resource *model.Resource) error {

	namespace, _ := util.ParseRepository(resource.Metadata.Repository.Name)
	ns, err := adapter.GetNamespace(namespace)
	if err != nil {
		//
	} else {
		if ns.Name == namespace {
			return nil
		}
	}

	url := fmt.Sprintf("%s/dockyard/v2/namespaces", adapter.Registry.URL)
	namespacebyte, err := json.Marshal(struct {
		Namespace string `json:"namespace"`
	}{Namespace: namespace})
	if err != nil {
		return err
	}

	r, err := http.NewRequest("POST", url, strings.NewReader(string(namespacebyte)))
	if err != nil {
		return err
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
		return err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("[%d][%s]", code, string(body))
	}
	return nil
}

// GetNamespace gets a namespace from Huawei SWR
func (adapter Adapter) GetNamespace(namespaceStr string) (*model.Namespace, error) {
	var namespace = &model.Namespace{
		Name:     "",
		Metadata: make(map[string]interface{}),
	}

	urls := fmt.Sprintf("%s/dockyard/v2/namespaces/%s", adapter.Registry.URL, namespaceStr)
	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return namespace, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")
	encodeAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", adapter.Registry.Credential.AccessKey, adapter.Registry.Credential.AccessSecret)))
	r.Header.Add("Authorization", "Basic "+encodeAuth)

	var client *http.Client
	if adapter.Registry.Insecure == true {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		client = &http.Client{}
	}
	resp, err := client.Do(r)
	if err != nil {
		return namespace, err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return namespace, fmt.Errorf("[%d][%s]", code, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return namespace, err
	}

	var namespaceData hwNamespace
	err = json.Unmarshal(body, &namespaceData)
	if err != nil {
		return namespace, err
	}

	namespace.Name = namespaceData.Name
	namespace.Metadata = namespaceData.metadata()

	return namespace, nil
}

// HealthCheck check health for huawei SWR
func (adapter Adapter) HealthCheck() (model.HealthStatus, error) {
	return model.Healthy, nil
}

// AdapterFactory is the factory for huawei adapter
func AdapterFactory(registry *model.Registry) (adapter.Adapter, error) {
	var adapter Adapter

	adapter.Registry = registry

	return adapter, nil

}

type hwNamespaceList struct {
	Namespace []hwNamespace `json:"namespaces"`
}

type hwNamespace struct {
	ID           int64  `json:"id" orm:"column(id)"`
	Name         string `json:"name"`
	CreatorName  string `json:"creator_name,omitempty"`
	DomainPublic int    `json:"-"`
	Auth         int    `json:"auth"`
	DomainName   string `json:"-"`
	UserCount    int64  `json:"user_count"`
	ImageCount   int64  `json:"image_count"`
}

func (ns hwNamespace) metadata() map[string]interface{} {
	var metadata = make(map[string]interface{})
	metadata["id"] = ns.ID
	metadata["creator_name"] = ns.CreatorName
	metadata["domain_public"] = ns.DomainPublic
	metadata["auth"] = ns.Auth
	metadata["domain_name"] = ns.DomainName
	metadata["user_count"] = ns.UserCount
	metadata["image_count"] = ns.ImageCount

	return metadata
}
