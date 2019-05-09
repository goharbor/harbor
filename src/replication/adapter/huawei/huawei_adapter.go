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

	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	err := adp.RegisterFactory(model.RegistryTypeHuawei, AdapterFactory)
	if err != nil {
		log.Errorf("failed to register factory for Huawei: %v", err)
		return
	}
	log.Infof("the factory of Huawei adapter was registered")
}

// Adapter is for images replications between harbor and Huawei image repository(SWR)
type adapter struct {
	*adp.DefaultImageRegistry
	registry *model.Registry
}

// Info gets info about Huawei SWR
func (a *adapter) Info() (*model.RegistryInfo, error) {
	registryInfo := model.RegistryInfo{
		Type:                     model.RegistryTypeHuawei,
		Description:              "Adapter for SWR -- The image registry of Huawei Cloud",
		SupportedResourceTypes:   []model.ResourceType{model.ResourceTypeImage},
		SupportedResourceFilters: []*model.FilterStyle{},
		SupportedTriggers:        []model.TriggerType{},
	}
	return &registryInfo, nil
}

// ListNamespaces lists namespaces from Huawei SWR with the provided query conditions.
func (a *adapter) ListNamespaces(query *model.NamespaceQuery) ([]*model.Namespace, error) {
	var namespaces []*model.Namespace

	urls := fmt.Sprintf("%s/dockyard/v2/visible/namespaces", a.registry.URL)

	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return namespaces, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")
	encodeAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", a.registry.Credential.AccessKey, a.registry.Credential.AccessSecret)))
	r.Header.Add("Authorization", "Basic "+encodeAuth)

	client := &http.Client{}
	if a.registry.Insecure == true {
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
func (a *adapter) ConvertResourceMetadata(resourceMetadata *model.ResourceMetadata, namespace *model.Namespace) (*model.ResourceMetadata, error) {
	metadata := &model.ResourceMetadata{
		Repository: resourceMetadata.Repository,
		Vtags:      resourceMetadata.Vtags,
		Labels:     resourceMetadata.Labels,
	}
	return metadata, nil
}

// PrepareForPush prepare for push to Huawei SWR
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	namespaces := map[string]struct{}{}
	for _, resource := range resources {
		paths := strings.Split(resource.Metadata.Repository.Name, "/")
		namespace := paths[0]
		ns, err := a.GetNamespace(namespace)
		if err != nil {
			return err
		}
		if ns != nil && ns.Name == namespace {
			continue
		}
		namespaces[namespace] = struct{}{}
	}

	url := fmt.Sprintf("%s/dockyard/v2/namespaces", a.registry.URL)
	client := &http.Client{
		Transport: util.GetHTTPTransport(a.registry.Insecure),
	}
	for namespace := range namespaces {
		namespacebyte, err := json.Marshal(struct {
			Namespace string `json:"namespace"`
		}{
			Namespace: namespace,
		})
		if err != nil {
			return err
		}

		r, err := http.NewRequest("POST", url, strings.NewReader(string(namespacebyte)))
		if err != nil {
			return err
		}

		r.Header.Add("content-type", "application/json; charset=utf-8")
		encodeAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", a.registry.Credential.AccessKey, a.registry.Credential.AccessSecret)))
		r.Header.Add("Authorization", "Basic "+encodeAuth)

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

		log.Debugf("namespace %s created", namespace)
	}
	return nil
}

// GetNamespace gets a namespace from Huawei SWR
func (a *adapter) GetNamespace(namespaceStr string) (*model.Namespace, error) {
	var namespace = &model.Namespace{
		Name:     "",
		Metadata: make(map[string]interface{}),
	}

	urls := fmt.Sprintf("%s/dockyard/v2/namespaces/%s", a.registry.URL, namespaceStr)
	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return namespace, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")
	encodeAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", a.registry.Credential.AccessKey, a.registry.Credential.AccessSecret)))
	r.Header.Add("Authorization", "Basic "+encodeAuth)

	var client *http.Client
	if a.registry.Insecure == true {
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
func (a *adapter) HealthCheck() (model.HealthStatus, error) {
	return model.Healthy, nil
}

// AdapterFactory is the factory for huawei adapter
func AdapterFactory(registry *model.Registry) (adp.Adapter, error) {
	reg, err := adp.NewDefaultImageRegistry(registry)
	if err != nil {
		return nil, err
	}
	return &adapter{
		registry:             registry,
		DefaultImageRegistry: reg,
	}, nil

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
