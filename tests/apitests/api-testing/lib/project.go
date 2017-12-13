package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vmware/harbor/tests/apitests/api-testing/client"
	"github.com/vmware/harbor/tests/apitests/api-testing/models"
)

//ProjectUtil : Util methods for project related
type ProjectUtil struct {
	rootURI       string
	testingClient *client.APIClient
}

//NewProjectUtil : Constructor
func NewProjectUtil(rootURI string, httpClient *client.APIClient) *ProjectUtil {
	if len(strings.TrimSpace(rootURI)) == 0 || httpClient == nil {
		return nil
	}

	return &ProjectUtil{
		rootURI:       rootURI,
		testingClient: httpClient,
	}
}

//GetProjects : Get projects
//If name specified, then only get the specified project
func (pu *ProjectUtil) GetProjects(name string) ([]models.ExistingProject, error) {
	url := pu.rootURI + "/api/projects"
	if len(strings.TrimSpace(name)) > 0 {
		url = url + "?name=" + name
	}
	data, err := pu.testingClient.Get(url)
	if err != nil {
		return nil, err
	}

	var pros []models.ExistingProject
	if err = json.Unmarshal(data, &pros); err != nil {
		return nil, err
	}

	return pros, nil
}

//GetProjectID : Get the project ID
//If no project existing with the name, then return -1
func (pu *ProjectUtil) GetProjectID(projectName string) int {
	pros, err := pu.GetProjects(projectName)
	if err != nil {
		return -1
	}

	if len(pros) == 0 {
		return -1
	}

	for _, pro := range pros {
		if pro.Name == projectName {
			return pro.ID
		}
	}

	return -1
}

//CreateProject :Create project
func (pu *ProjectUtil) CreateProject(projectName string, accessLevel bool) error {
	if len(strings.TrimSpace(projectName)) == 0 {
		return errors.New("Empty project name for creating")
	}

	p := models.Project{
		Name: projectName,
		Metadata: &models.Metadata{
			AccessLevel: fmt.Sprintf("%v", accessLevel),
		},
	}

	body, err := json.Marshal(&p)
	if err != nil {
		return err
	}

	url := pu.rootURI + "/api/projects"
	if err = pu.testingClient.Post(url, body); err != nil {
		return err
	}

	return nil
}

//DeleteProject : Delete project
func (pu *ProjectUtil) DeleteProject(projectName string) error {
	if len(strings.TrimSpace(projectName)) == 0 {
		return errors.New("Empty project name for deleting")
	}

	pid := pu.GetProjectID(projectName)
	if pid == -1 {
		return errors.New("Failed to get project ID")
	}

	url := fmt.Sprintf("%s%s%d", pu.rootURI, "/api/projects/", pid)

	if err := pu.testingClient.Delete(url); err != nil {
		return err
	}

	return nil
}

//AssignRole : Assign role to user
func (pu *ProjectUtil) AssignRole(projectName, username string) error {
	if len(strings.TrimSpace(projectName)) == 0 ||
		len(strings.TrimSpace(username)) == 0 {
		return errors.New("Project name and username are required for assigning role")
	}

	pid := pu.GetProjectID(projectName)
	if pid == -1 {
		return fmt.Errorf("Failed to get project ID with name %s", projectName)
	}

	m := models.Member{
		UserName: username,
		Roles:    []int{2},
	}

	body, err := json.Marshal(&m)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s%s%d%s", pu.rootURI, "/api/projects/", pid, "/members")
	if err := pu.testingClient.Post(url, body); err != nil {
		return err
	}

	return nil
}

//RevokeRole : RevokeRole role from user
func (pu *ProjectUtil) RevokeRole(projectName string, uid int) error {
	if len(strings.TrimSpace(projectName)) == 0 {
		return errors.New("Project name is required for revoking role")
	}

	if uid == 0 {
		return errors.New("User ID is required for revoking role")
	}

	pid := pu.GetProjectID(projectName)
	if pid == -1 {
		return fmt.Errorf("Failed to get project ID with name %s", projectName)
	}

	url := fmt.Sprintf("%s%s%d%s%d", pu.rootURI, "/api/projects/", pid, "/members/", uid)
	if err := pu.testingClient.Delete(url); err != nil {
		return err
	}

	return nil
}
