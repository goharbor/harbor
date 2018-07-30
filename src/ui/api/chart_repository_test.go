package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/ui/promgr/metamgr"
)

//Test the URL rewrite function
func TestURLRewrite(t *testing.T) {
	chartAPI := &ChartRepositoryAPI{}
	req, err := createRequest(http.MethodGet, "/api/chartrepo/health")
	if err != nil {
		t.Fatal(err)
	}
	chartAPI.rewriteURLPath(req)
	if req.URL.Path != "/health" {
		t.Fatalf("Expect url format %s but got %s", "/health", req.URL.Path)
	}

	req, err = createRequest(http.MethodGet, "/api/chartrepo/library/charts")
	if err != nil {
		t.Fatal(err)
	}
	chartAPI.rewriteURLPath(req)
	if req.URL.Path != "/api/library/charts" {
		t.Fatalf("Expect url format %s but got %s", "/api/library/charts", req.URL.Path)
	}

	req, err = createRequest(http.MethodPost, "/api/chartrepo/charts")
	if err != nil {
		t.Fatal(err)
	}
	chartAPI.rewriteURLPath(req)
	if req.URL.Path != "/api/library/charts" {
		t.Fatalf("Expect url format %s but got %s", "/api/library/charts", req.URL.Path)
	}

	req, err = createRequest(http.MethodGet, "/chartrepo/library/index.yaml")
	if err != nil {
		t.Fatal(err)
	}
	chartAPI.rewriteURLPath(req)
	if req.URL.Path != "/library/index.yaml" {
		t.Fatalf("Expect url format %s but got %s", "/library/index.yaml", req.URL.Path)
	}
}

//Test access checking
func TestRequireAccess(t *testing.T) {
	chartAPI := &ChartRepositoryAPI{}
	chartAPI.SecurityCtx = &mockSecurityContext{}

	ns := "library"
	if !chartAPI.requireAccess(ns, accessLevelPublic) {
		t.Fatal("expect true result (public access level is granted) but got false")
	}
	if !chartAPI.requireAccess(ns, accessLevelAll) {
		t.Fatal("expect true result (admin has all perm) but got false")
	}
	if !chartAPI.requireAccess(ns, accessLevelRead) {
		t.Fatal("expect true result (admin has read perm) but got false")
	}
	if !chartAPI.requireAccess(ns, accessLevelWrite) {
		t.Fatal("expect true result (admin has write perm) but got false")
	}
	if !chartAPI.requireAccess(ns, accessLevelSystem) {
		t.Fatal("expect true result (admin has system perm) but got false")
	}
}

func TestIsMultipartFormData(t *testing.T) {
	req, err := createRequest(http.MethodPost, "/api/chartrepo/charts")
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(headerContentType, "application/json")
	if isMultipartFormData(req) {
		t.Fatal("expect false result but got true")
	}

	req.Header.Set(headerContentType, contentTypeMultipart)
	if !isMultipartFormData(req) {
		t.Fatalf("expect %s result but got %s", contentTypeMultipart, req.Header.Get(headerContentType))
	}
}

//Test namespace cheking
func TestRequireNamespace(t *testing.T) {
	chartAPI := &ChartRepositoryAPI{}
	chartAPI.ProjectMgr = &mockProjectManager{}

	if !chartAPI.requireNamespace("library") {
		t.Fatal("expect namespace 'library' existing but got false")
	}
}

func createRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.RequestURI = url

	return req, nil
}

//Mock project manager
type mockProjectManager struct{}

func (mpm *mockProjectManager) Get(projectIDOrName interface{}) (*models.Project, error) {
	return nil, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Create(*models.Project) (int64, error) {
	return -1, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Delete(projectIDOrName interface{}) error {
	return errors.New("Not implemented")
}

func (mpm *mockProjectManager) Update(projectIDOrName interface{}, project *models.Project) error {
	return errors.New("Not implemented")
}
func (mpm *mockProjectManager) List(query *models.ProjectQueryParam) (*models.ProjectQueryResult, error) {
	results := &models.ProjectQueryResult{
		Total:    2,
		Projects: make([]*models.Project, 0),
	}

	results.Projects = append(results.Projects, &models.Project{ProjectID: 0, Name: "repo1"})
	results.Projects = append(results.Projects, &models.Project{ProjectID: 1, Name: "repo2"})

	return results, nil
}

func (mpm *mockProjectManager) IsPublic(projectIDOrName interface{}) (bool, error) {
	return false, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Exists(projectIDOrName interface{}) (bool, error) {
	if projectIDOrName == nil {
		return false, errors.New("nil projectIDOrName")
	}

	if ns, ok := projectIDOrName.(string); ok {
		if ns == "library" {
			return true, nil
		}

		return false, nil
	}

	return false, errors.New("unknown type of projectIDOrName")
}

// get all public project
func (mpm *mockProjectManager) GetPublic() ([]*models.Project, error) {
	return nil, errors.New("Not implemented")
}

// if the project manager uses a metadata manager, return it, otherwise return nil
func (mpm *mockProjectManager) GetMetadataManager() metamgr.ProjectMetadataManager {
	return nil
}

//mock security context
type mockSecurityContext struct{}

// IsAuthenticated returns whether the context has been authenticated or not
func (msc *mockSecurityContext) IsAuthenticated() bool {
	return true
}

// GetUsername returns the username of user related to the context
func (msc *mockSecurityContext) GetUsername() string {
	return "amdin"
}

// IsSysAdmin returns whether the user is system admin
func (msc *mockSecurityContext) IsSysAdmin() bool {
	return true
}

// IsSolutionUser returns whether the user is solution user
func (msc *mockSecurityContext) IsSolutionUser() bool {
	return false
}

// HasReadPerm returns whether the user has read permission to the project
func (msc *mockSecurityContext) HasReadPerm(projectIDOrName interface{}) bool {
	if projectIDOrName == nil {
		return false
	}

	if ns, ok := projectIDOrName.(string); ok {
		if ns == "library" {
			return true
		}
	}

	return false
}

// HasWritePerm returns whether the user has write permission to the project
func (msc *mockSecurityContext) HasWritePerm(projectIDOrName interface{}) bool {
	if projectIDOrName == nil {
		return false
	}

	if ns, ok := projectIDOrName.(string); ok {
		if ns == "library" {
			return true
		}
	}

	return false
}

// HasAllPerm returns whether the user has all permissions to the project
func (msc *mockSecurityContext) HasAllPerm(projectIDOrName interface{}) bool {
	return msc.HasReadPerm(projectIDOrName) && msc.HasWritePerm(projectIDOrName)
}

//Get current user's all project
func (msc *mockSecurityContext) GetMyProjects() ([]*models.Project, error) {
	return []*models.Project{&models.Project{ProjectID: 0, Name: "repo1"}}, nil
}

//Get user's role in provided project
func (msc *mockSecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	return []int{0, 1, 2, 3}
}
