package jfrog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	err := adp.RegisterFactory(model.RegistryTypeJfrogArtifactory, new(factory))
	if err != nil {
		log.Errorf("failed to register factory for jfrog artifactory: %v", err)
		return
	}
	log.Infof("the factory of jfrog artifactory adapter was registered")
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

// Adapter is for images replications between harbor and jfrog artifactory image repository
type adapter struct {
	*native.Adapter
	registry *model.Registry
	client   *common_http.Client
}

var _ adp.Adapter = (*adapter)(nil)

// Info gets info about jfrog artifactory adapter
func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	info = &model.RegistryInfo{
		Type: model.RegistryTypeJfrogArtifactory,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeImage,
		},
		SupportedResourceFilters: []*model.FilterStyle{
			{
				Type:  model.FilterTypeName,
				Style: model.FilterStyleTypeText,
			},
			{
				Type:  model.FilterTypeTag,
				Style: model.FilterStyleTypeText,
			},
		},
		SupportedTriggers: []model.TriggerType{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}
	return
}

func newAdapter(registry *model.Registry) (adp.Adapter, error) {
	var (
		modifiers = []modifier.Modifier{}
	)
	if registry.Credential != nil {
		modifiers = append(modifiers, basic.NewAuthorizer(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret))
	}

	return &adapter{
		Adapter:  native.NewAdapter(registry),
		registry: registry,
		client: common_http.NewClient(
			&http.Client{
				Transport: util.GetHTTPTransport(registry.Insecure),
			},
			modifiers...,
		),
	}, nil

}

// PrepareForPush creates local docker repository in jfrog artifactory
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	var namespaces []string
	for _, resource := range resources {
		if resource == nil {
			return errors.New("the resource cannot be null")
		}
		if resource.Metadata == nil {
			return errors.New("the metadata of resource cannot be null")
		}
		if resource.Metadata.Repository == nil {
			return errors.New("the namespace of resource cannot be null")
		}
		if len(resource.Metadata.Repository.Name) == 0 {
			return errors.New("the name of namespace cannot be null")
		}
		path := strings.Split(resource.Metadata.Repository.Name, "/")
		if len(path) > 0 {
			namespaces = append(namespaces, path[0])
		}
	}

	repositories, err := a.getLocalRepositories()
	if err != nil {
		log.Errorf("Get local repositories error: %v", err)
		return err
	}

	existedRepositories := make(map[string]struct{})
	for _, repo := range repositories {
		existedRepositories[repo.Key] = struct{}{}
	}

	for _, namespace := range namespaces {
		if _, ok := existedRepositories[namespace]; ok {
			log.Debugf("Namespace %s already existed in remote, skip create it", namespace)
		} else {
			err = a.createNamespace(namespace)
			if err != nil {
				log.Errorf("Create Namespace %s error: %v", namespace, err)
				return err
			}
		}
	}

	return nil
}

func (a *adapter) getLocalRepositories() ([]*repository, error) {
	var repositories []*repository
	url := fmt.Sprintf("%s/artifactory/api/repositories?type=local&packageType=docker", a.registry.URL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return repositories, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return repositories, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return repositories, err
	}

	err = json.Unmarshal(body, &repositories)
	return repositories, err
}

// create repository with docker local type
// this operation needs admin
func (a *adapter) createNamespace(namespace string) error {
	ns := newDefaultDockerLocalRepository(namespace)
	body, err := json.Marshal(ns)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/artifactory/api/repositories/%s", a.registry.URL, namespace)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}

// PushBlob can not use naive PushBlob due to MonolithicUpload, Jfrog now just support push by chunk
// related issue: https://www.jfrog.com/jira/browse/RTFACT-19344
func (a *adapter) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	location, err := a.preparePushBlob(repository)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v2/%s/blobs/uploads/%s", a.registry.URL, repository, location)
	req, err := http.NewRequest(http.MethodPatch, url, blob)
	if err != nil {
		return err
	}
	rangeSize := strconv.Itoa(int(size))
	req.Header.Set("Content-Length", rangeSize)
	req.Header.Set("Content-Range", fmt.Sprintf("0-%s", rangeSize))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return a.ackPushBlob(repository, digest, location, rangeSize)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}

func (a *adapter) preparePushBlob(repository string) (string, error) {
	url := fmt.Sprintf("%s/v2/%s/blobs/uploads/", a.registry.URL, repository)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set(http.CanonicalHeaderKey("Content-Length"), "0")
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return resp.Header.Get(http.CanonicalHeaderKey("Docker-Upload-Uuid")), nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}

	return "", err
}

func (a *adapter) ackPushBlob(repository, digest, location, size string) error {
	url := fmt.Sprintf("%s/v2/%s/blobs/uploads/%s?digest=%s", a.registry.URL, repository, location, digest)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}

	return err
}
