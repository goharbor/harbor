//Package HarborAPI
//These APIs provide services for manipulating Harbor project.
package HarborAPI

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dghubble/sling"
)

type HarborAPI struct {
	basePath string
}

func NewHarborAPI() *HarborAPI {
	return &HarborAPI{
		basePath: "http://localhost",
	}
}

func NewHarborAPIWithBasePath(basePath string) *HarborAPI {
	return &HarborAPI{
		basePath: basePath,
	}
}

type UsrInfo struct {
	Name   string
	Passwd string
}

//Search for projects and repositories
//Implementation Notes
//The Search endpoint returns information about the projects and repositories
//offered at public status or related to the current logged in user.
//The response includes the project and repository list in a proper display order.
//@param q Search parameter for project and repository name.
//@return []Search
//func (a HarborAPI) SearchGet (q string) (Search, error) {
func (a HarborAPI) SearchGet(q string) (Search, error) {

        _sling := sling.New().Get(a.basePath)

	// create path and map variables
	path := "/api/search"

	_sling = _sling.Path(path)

	type QueryParams struct {
		Query string `url:"q"`
	}

	_sling = _sling.QueryStruct(&QueryParams{q})

	// accept header
	accepts := []string{"application/json", "text/plain"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	req, err := _sling.Request()
        client := &http.Client{}
	httpResponse, err := client.Do(req)
	defer httpResponse.Body.Close()
	
        body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		// handle error
	}
	
        var successPayload = new(Search)
	err = json.Unmarshal(body, &successPayload)
	return *successPayload, err
}

//Create a new project.
//Implementation Notes
//This endpoint is for user to create a new project.
//@param project New created project.
//@return void
//func (a HarborAPI) ProjectsPost (prjUsr UsrInfo, project Project) (int, error) {
func (a HarborAPI) ProjectsPost(prjUsr UsrInfo, project Project) (int, error) {

	_sling := sling.New().Post(a.basePath)

	// create path and map variables
	path := "/api/projects"

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json", "text/plain"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	// body params
	_sling = _sling.BodyJSON(project)

	req, err := _sling.Request()
	req.SetBasicAuth(prjUsr.Name, prjUsr.Passwd)

        client := &http.Client{}
        httpResponse, err := client.Do(req)
        defer httpResponse.Body.Close()

	return httpResponse.StatusCode, err
}
