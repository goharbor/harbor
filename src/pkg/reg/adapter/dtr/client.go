package dtr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Client is a client to interact with DTR
type Client struct {
	client   *common_http.Client
	url      string
	username string
	password string
}

// NewClient creates a new DTR client.
func NewClient(registry *model.Registry) *Client {

	client := &Client{
		url:      registry.URL,
		username: registry.Credential.AccessKey,
		password: registry.Credential.AccessSecret,
		client: common_http.NewClient(
			&http.Client{
				Transport: common_http.GetHTTPTransport(common_http.WithInsecure(registry.Insecure)),
			}),
	}
	return client
}

// getAndIteratePagination will iterator over a paginated response from DTR
func (c *Client) getAndIteratePagination(endpoint string, v interface{}) error {
	urlAPI, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v should be a pointer to a slice")
	}
	elemType := rv.Elem().Type()
	if elemType.Kind() != reflect.Slice {
		return errors.New("v should be a pointer to a slice")
	}

	resources := reflect.Indirect(reflect.New(elemType))
	for len(endpoint) > 0 {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(c.username, c.password)
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return &common_http.Error{
				Code:    resp.StatusCode,
				Message: string(data),
			}
		}

		res := reflect.New(elemType)
		if err = json.Unmarshal(data, res.Interface()); err != nil {
			log.Errorf("Failed to parse json response: %v", string(data))
			return err
		}
		resources = reflect.AppendSlice(resources, reflect.Indirect(res))
		endpoint = ""

		nextPage := resp.Header.Get("X-Next-Page-Start")
		if len(nextPage) > 0 {
			query := urlAPI.Query()
			query.Set("pageStart", nextPage)
			endpoint = urlAPI.Scheme + "://" + urlAPI.Host + urlAPI.Path + "?" + query.Encode()
		}
	}
	rv.Elem().Set(resources)
	return nil
}

// getRepositories returns a list of repositories in DTR
func (c *Client) getRepositories() ([]*model.Repository, error) {
	var repositories []Repository
	var dtrRepositories Repositories

	endpoint := fmt.Sprintf("%s/api/v0/repositories?pageSize=100", c.url)
	urlAPI, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	for len(endpoint) > 0 {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(c.username, c.password)
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, &common_http.Error{
				Code:    resp.StatusCode,
				Message: string(data),
			}
		}

		if err = json.Unmarshal(data, &dtrRepositories); err != nil {
			log.Errorf("Failed to parse json response")
			log.Errorf("%v", err)
			log.Errorf("%s", string(data))

			return nil, err
		}

		// merge the arrays
		repositories = append(repositories, dtrRepositories.Repositories...)
		endpoint = ""

		nextPage := resp.Header.Get("X-Next-Page-Start")
		if len(nextPage) > 0 {
			query := urlAPI.Query()
			query.Set("pageStart", nextPage)
			endpoint = urlAPI.Scheme + "://" + urlAPI.Host + urlAPI.Path + "?" + query.Encode()
		}
	}

	result := []*model.Repository{}

	for _, repository := range repositories {
		log.Debugf("Processing DTR repo %s", repository.Name)
		result = append(result, &model.Repository{
			Name: fmt.Sprintf("%s/%s", repository.Namespace, repository.Name),
		})
	}
	return result, nil
}

// getTags looks up a repositories tags in DTR
func (c *Client) getTags(repository string) ([]string, error) {
	var tags []*Tag
	// This assumes repository is of form namespace/repo
	urlAPI := fmt.Sprintf("%s/api/v0/repositories/%s/tags?pageSize=100", c.url, repository)
	log.Debugf("Looking up tags for %s at %s", repository, urlAPI)
	if err := c.getAndIteratePagination(urlAPI, &tags); err != nil {
		log.Debugf("Failed looking up tags for %s at %s", repository, urlAPI)
		return nil, err
	}

	var result []string
	for _, tag := range tags {
		result = append(result, tag.Name)
	}
	return result, nil
}

// getNamespaces returns DTR namespaces.  DTR also calles these orgs and accounts depending on where you look
func (c *Client) getNamespaces() ([]Account, error) {
	var accounts []Account
	var response Accounts

	endpoint := fmt.Sprintf("%s/enzi/v0/accounts?limit=100", c.url)
	urlAPI, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	for len(endpoint) > 0 {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(c.username, c.password)
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, &common_http.Error{
				Code:    resp.StatusCode,
				Message: string(data),
			}
		}

		if err = json.Unmarshal(data, &response); err != nil {
			log.Errorf("Failed to parse json response")
			log.Errorf("%v", err)
			log.Errorf("%s", string(data))

			return nil, err
		}

		accounts = append(accounts, response.Accounts...)
		endpoint = ""

		nextPage := resp.Header.Get("X-Next-Page-Start")
		if len(nextPage) > 0 {
			query := urlAPI.Query()
			query.Set("start", nextPage)
			endpoint = urlAPI.Scheme + "://" + urlAPI.Host + urlAPI.Path + "?" + query.Encode()
		}
	}

	return accounts, nil
}

// createRepository creates a repository in DTR.  The namespace/org/account must already exist.
func (c *Client) createRepository(repository string) error {
	var namespace string
	var repositoryName string

	path := strings.Split(repository, "/")
	if len(path) > 1 {
		namespace = path[0]
		repositoryName = path[1]
	} else {
		return errors.New("repository did not contain a namespace")
	}

	repo := newDefaultDTRRepository(repositoryName)
	body, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	urlAPI := fmt.Sprintf("%s/api/v0/repositories/%s", c.url, namespace)
	log.Debugf("Creating repo %s in DTR at %s", repositoryName, urlAPI)
	req, err := http.NewRequest(http.MethodPost, urlAPI, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
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

	return &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}

// createNamespace creates a namespace in DTR
// This actually hits the enzi API which appears to map to the UCP
// accounts API.  The DTR v0 api has no official way to create a
// namespace as of 2.7.1
// this operation needs admin access
func (c *Client) createNamespace(namespace string) error {
	ns := newDefaultDTRNamespace(namespace)
	body, err := json.Marshal(ns)
	if err != nil {
		return err
	}

	urlAPI := fmt.Sprintf("%s/enzi/v0/accounts", c.url)
	log.Debugf("Creating namespace %s in DTR at %s", namespace, urlAPI)
	req, err := http.NewRequest(http.MethodPost, urlAPI, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
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

	return &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}
