package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/goharbor/harbor/src/chartserver"
	hlog "github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/ui/config"
)

const (
	namespaceParam          = ":repo"
	defaultRepo             = "library"
	rootUploadingEndpoint   = "/api/chartrepo/charts"
	rootIndexEndpoint       = "/chartrepo/index.yaml"
	chartRepoHealthEndpoint = "/api/chartrepo/health"

	accessLevelPublic = iota
	accessLevelRead
	accessLevelWrite
	accessLevelAll
	accessLevelSystem

	formFieldNameForChart = "chart"
	formFiledNameForProv  = "prov"
	headerContentType     = "Content-Type"
	contentTypeMultipart  = "multipart/form-data"
)

//chartController is a singleton instance
var chartController *chartserver.Controller

//ChartRepositoryAPI provides related API handlers for the chart repository APIs
type ChartRepositoryAPI struct {
	//The base controller to provide common utilities
	BaseController

	//Keep the namespace if existing
	namespace string
}

//Prepare something for the following actions
func (cra *ChartRepositoryAPI) Prepare() {
	//Call super prepare method
	cra.BaseController.Prepare()

	//Try to extract namespace for parameter of path
	//It may not exist
	cra.namespace = strings.TrimSpace(cra.GetStringFromPath(namespaceParam))

	//Check the existence of namespace
	//Exclude the following URI
	// -/index.yaml
	// -/api/chartserver/health
	incomingURI := cra.Ctx.Request.RequestURI
	if incomingURI == rootUploadingEndpoint {
		//Forward to the default repository
		cra.namespace = defaultRepo
	}

	if incomingURI != rootIndexEndpoint &&
		incomingURI != chartRepoHealthEndpoint {
		if !cra.requireNamespace(cra.namespace) {
			return
		}
	}

	//Rewrite URL path
	cra.rewriteURLPath(cra.Ctx.Request)
}

//GetHealthStatus handles GET /api/chartserver/health
func (cra *ChartRepositoryAPI) GetHealthStatus() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelSystem) {
		return
	}

	chartController.GetBaseHandler().GetHealthStatus(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//GetIndexByRepo handles GET /:repo/index.yaml
func (cra *ChartRepositoryAPI) GetIndexByRepo() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	chartController.GetRepositoryHandler().GetIndexFileWithNS(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//GetIndex handles GET /index.yaml
func (cra *ChartRepositoryAPI) GetIndex() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelSystem) {
		return
	}

	chartController.GetRepositoryHandler().GetIndexFile(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//DownloadChart handles GET /:repo/charts/:filename
func (cra *ChartRepositoryAPI) DownloadChart() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	chartController.GetRepositoryHandler().DownloadChartObject(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//ListCharts handles GET /api/:repo/charts
func (cra *ChartRepositoryAPI) ListCharts() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	chartController.GetManipulationHandler().ListCharts(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//ListChartVersions GET /api/:repo/charts/:name
func (cra *ChartRepositoryAPI) ListChartVersions() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	chartController.GetManipulationHandler().GetChart(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//GetChartVersion handles GET /api/:repo/charts/:name/:version
func (cra *ChartRepositoryAPI) GetChartVersion() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	//Let's pass the namespace via the context of request
	req := cra.Ctx.Request
	*req = *(req.WithContext(context.WithValue(req.Context(), chartserver.NamespaceContextKey, cra.namespace)))

	chartController.GetManipulationHandler().GetChartVersion(cra.Ctx.ResponseWriter, req)
}

//DeleteChartVersion handles DELETE /api/:repo/charts/:name/:version
func (cra *ChartRepositoryAPI) DeleteChartVersion() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelAll) {
		return
	}

	chartController.GetManipulationHandler().DeleteChartVersion(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//UploadChartVersion handles POST /api/:repo/charts
func (cra *ChartRepositoryAPI) UploadChartVersion() {
	hlog.Debugf("Header of request of uploading chart: %#v, content-len=%d", cra.Ctx.Request.Header, cra.Ctx.Request.ContentLength)

	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelWrite) {
		return
	}

	//Rewrite file content if the content type is "multipart/form-data"
	if isMultipartFormData(cra.Ctx.Request) {
		formFiles := make([]formFile, 0)
		formFiles = append(formFiles,
			formFile{
				formField: formFieldNameForChart,
				mustHave:  true,
			},
			formFile{
				formField: formFiledNameForProv,
			})
		if err := cra.rewriteFileContent(formFiles, cra.Ctx.Request); err != nil {
			chartserver.WriteInternalError(cra.Ctx.ResponseWriter, err)
			return
		}
	}

	chartController.GetManipulationHandler().UploadChartVersion(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//UploadChartProvFile handles POST /api/:repo/prov
func (cra *ChartRepositoryAPI) UploadChartProvFile() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelWrite) {
		return
	}

	//Rewrite file content if the content type is "multipart/form-data"
	if isMultipartFormData(cra.Ctx.Request) {
		formFiles := make([]formFile, 0)
		formFiles = append(formFiles,
			formFile{
				formField: formFiledNameForProv,
				mustHave:  true,
			})
		if err := cra.rewriteFileContent(formFiles, cra.Ctx.Request); err != nil {
			chartserver.WriteInternalError(cra.Ctx.ResponseWriter, err)
			return
		}
	}

	chartController.GetManipulationHandler().UploadProvenanceFile(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//Rewrite the incoming URL with the right backend URL pattern
//Remove 'chartrepo' from the endpoints of manipulation API
//Remove 'chartrepo' from the endpoints of repository services
func (cra *ChartRepositoryAPI) rewriteURLPath(req *http.Request) {
	incomingURLPath := req.RequestURI

	defer func() {
		hlog.Debugf("Incoming URL '%s' is rewritten to '%s'", incomingURLPath, req.URL.String())
	}()

	//Health check endpoint
	if incomingURLPath == chartRepoHealthEndpoint {
		req.URL.Path = "/health"
		return
	}

	//Root uploading endpoint
	if incomingURLPath == rootUploadingEndpoint {
		req.URL.Path = strings.Replace(incomingURLPath, "chartrepo", defaultRepo, 1)
		return
	}

	//Repository endpoints
	if strings.HasPrefix(incomingURLPath, "/chartrepo") {
		req.URL.Path = strings.TrimPrefix(incomingURLPath, "/chartrepo")
		return
	}

	//API endpoints
	if strings.HasPrefix(incomingURLPath, "/api/chartrepo") {
		req.URL.Path = strings.Replace(incomingURLPath, "/chartrepo", "", 1)
		return
	}
}

//Check if there exists a valid namespace
//Return true if it does
//Return false if it does not
func (cra *ChartRepositoryAPI) requireNamespace(namespace string) bool {
	//Actually, never should be like this
	if len(namespace) == 0 {
		cra.HandleBadRequest(":repo should be in the request URL")
		return false
	}

	existsing, err := cra.ProjectMgr.Exists(namespace)
	if err != nil {
		//Check failed with error
		cra.renderError(http.StatusInternalServerError, fmt.Sprintf("failed to check existence of namespace %s with error: %s", namespace, err.Error()))
		return false
	}

	//Not existing
	if !existsing {
		cra.renderError(http.StatusBadRequest, fmt.Sprintf("namespace %s is not existing", namespace))
		return false
	}

	return true
}

//Check if the related access match the expected requirement
//If with right access, return true
//If without right access, return false
func (cra *ChartRepositoryAPI) requireAccess(namespace string, accessLevel uint) bool {
	if accessLevel == accessLevelPublic {
		return true //do nothing
	}

	theLevel := accessLevel
	//If repo is empty, system admin role must be required
	if len(namespace) == 0 {
		theLevel = accessLevelSystem
	}

	var err error

	switch theLevel {
	//Should be system admin role
	case accessLevelSystem:
		if !cra.SecurityCtx.IsSysAdmin() {
			err = errors.New("permission denied: system admin role is required")
		}
	case accessLevelAll:
		if !cra.SecurityCtx.HasAllPerm(namespace) {
			err = errors.New("permission denied: project admin or higher role is required")
		}
	case accessLevelWrite:
		if !cra.SecurityCtx.HasWritePerm(namespace) {
			err = errors.New("permission denied: developer or higher role is required")
		}
	case accessLevelRead:
		if !cra.SecurityCtx.HasReadPerm(namespace) {
			err = errors.New("permission denied: guest or higher role is required")
		}
	default:
		//access rejected for invalid scope
		cra.renderError(http.StatusForbidden, "unrecognized access scope")
		return false
	}

	//Access is not granted, check if user has authenticated
	if err != nil {
		//Unauthenticated, return 401
		if !cra.SecurityCtx.IsAuthenticated() {
			cra.renderError(http.StatusUnauthorized, "Unauthorized")
			return false
		}

		//Authenticated, return 403
		cra.renderError(http.StatusForbidden, err.Error())
		return false
	}

	return true
}

//write error message with unified format
func (cra *ChartRepositoryAPI) renderError(code int, text string) {
	chartserver.WriteError(cra.Ctx.ResponseWriter, code, errors.New(text))
}

//formFile is used to represent the uploaded files in the form
type formFile struct {
	//form field key contains the form file
	formField string

	//flag to indicate if the file identified by the 'formField'
	//must exist
	mustHave bool
}

//If the files are uploaded with multipart/form-data mimetype, beego will extract the data
//from the request automatically. Then the request passed to the backend server with proxying
//way will have empty content.
//This method will refill the requests with file content.
func (cra *ChartRepositoryAPI) rewriteFileContent(files []formFile, request *http.Request) error {
	if len(files) == 0 {
		return nil //no files, early return
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		if err := w.Close(); err != nil {
			//Just log it
			hlog.Errorf("Failed to defer close multipart writer with error: %s", err.Error())
		}
	}()

	//Process files by key one by one
	for _, f := range files {
		mFile, mHeader, err := cra.GetFile(f.formField)
		//Handle error case by case
		if err != nil {
			formatedErr := fmt.Errorf("Get file content with multipart header from key '%s' failed with error: %s", f.formField, err.Error())
			if f.mustHave || err != http.ErrMissingFile {
				return formatedErr
			}

			//Error can be ignored, just log it
			hlog.Warning(formatedErr.Error())
			continue
		}

		fw, err := w.CreateFormFile(f.formField, mHeader.Filename)
		if err != nil {
			return fmt.Errorf("Create form file with multipart header failed with error: %s", err.Error())
		}

		_, err = io.Copy(fw, mFile)
		if err != nil {
			return fmt.Errorf("Copy file stream in multipart form data failed with error: %s", err.Error())
		}
	}

	request.Header.Set(headerContentType, w.FormDataContentType())
	request.ContentLength = -1
	request.Body = ioutil.NopCloser(&body)

	return nil
}

//Initialize the chart service controller
func initializeChartController() (*chartserver.Controller, error) {
	addr, err := config.GetChartMuseumEndpoint()
	if err != nil {
		return nil, fmt.Errorf("Failed to get the endpoint URL of chart storage server: %s", err.Error())
	}

	addr = strings.TrimSuffix(addr, "/")
	url, err := url.Parse(addr)
	if err != nil {
		return nil, errors.New("Endpoint URL of chart storage server is malformed")
	}

	controller, err := chartserver.NewController(url)
	if err != nil {
		return nil, errors.New("Failed to initialize chart API controller")
	}

	hlog.Debugf("Chart storage server is set to %s", url.String())
	hlog.Info("API controller for chart repository server is successfully initialized")

	return controller, nil
}

//Check if the request content type is "multipart/form-data"
func isMultipartFormData(req *http.Request) bool {
	return strings.Contains(req.Header.Get(headerContentType), contentTypeMultipart)
}
