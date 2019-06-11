package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/testing/apitests/api-testing/client"
	"github.com/goharbor/harbor/src/testing/apitests/api-testing/models"
)

// ImageUtil : For repository and tag functions
type ImageUtil struct {
	rootURI       string
	testingClient *client.APIClient
}

// NewImageUtil : Constructor
func NewImageUtil(rootURI string, httpClient *client.APIClient) *ImageUtil {
	if len(strings.TrimSpace(rootURI)) == 0 || httpClient == nil {
		return nil
	}

	return &ImageUtil{
		rootURI:       rootURI,
		testingClient: httpClient,
	}
}

// DeleteRepo : Delete repo
func (iu *ImageUtil) DeleteRepo(repoName string) error {
	if len(strings.TrimSpace(repoName)) == 0 {
		return errors.New("Empty repo name for deleting")
	}

	url := fmt.Sprintf("%s%s%s", iu.rootURI, "/api/repositories/", repoName)
	if err := iu.testingClient.Delete(url); err != nil {
		return err
	}

	return nil
}

// ScanTag :Scan a tag
func (iu *ImageUtil) ScanTag(repoName string, tagName string) error {
	if len(strings.TrimSpace(repoName)) == 0 {
		return errors.New("Empty repo name for scanning")
	}

	if len(strings.TrimSpace(tagName)) == 0 {
		return errors.New("Empty tag name for scanning")
	}

	url := fmt.Sprintf("%s%s%s%s%s%s", iu.rootURI, "/api/repositories/", repoName, "/tags/", tagName, "/scan")
	if err := iu.testingClient.Post(url, nil); err != nil {
		return err
	}

	tk := time.NewTicker(1 * time.Second)
	defer tk.Stop()
	done := make(chan bool)
	errchan := make(chan error)
	url = fmt.Sprintf("%s%s%s%s%s", iu.rootURI, "/api/repositories/", repoName, "/tags/", tagName)
	go func() {
		for range tk.C {
			data, err := iu.testingClient.Get(url)
			if err != nil {
				errchan <- err
				return
			}
			var tag models.Tag
			if err = json.Unmarshal(data, &tag); err != nil {
				errchan <- err
				return
			}

			if tag.ScanOverview != nil && tag.ScanOverview.Status == "finished" {
				done <- true
			}
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(20 * time.Second):
		return errors.New("Scan timeout after 30 seconds")
	}
}

// GetRepos : Get repos in the project
func (iu *ImageUtil) GetRepos(projectName string) ([]models.Repository, error) {
	if len(strings.TrimSpace(projectName)) == 0 {
		return nil, errors.New("Empty project name for getting repos")
	}

	proUtil := NewProjectUtil(iu.rootURI, iu.testingClient)
	pid := proUtil.GetProjectID(projectName)
	if pid == -1 {
		return nil, fmt.Errorf("Failed to get project ID with name %s", projectName)
	}

	url := fmt.Sprintf("%s%s%d", iu.rootURI, "/api/repositories?project_id=", pid)
	data, err := iu.testingClient.Get(url)
	if err != nil {
		return nil, err
	}

	var repos []models.Repository
	if err = json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// GetTags : Get tags
func (iu *ImageUtil) GetTags(repoName string) ([]models.Tag, error) {
	if len(strings.TrimSpace(repoName)) == 0 {
		return nil, errors.New("Empty repository name for getting tags")
	}

	url := fmt.Sprintf("%s%s%s%s", iu.rootURI, "/api/repositories/", repoName, "/tags")
	tagData, err := iu.testingClient.Get(url)
	if err != nil {
		return nil, err
	}

	var tags []models.Tag
	if err = json.Unmarshal(tagData, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}
