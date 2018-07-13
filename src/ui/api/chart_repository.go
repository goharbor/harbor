package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/vmware/harbor/src/chartserver"
	hlog "github.com/vmware/harbor/src/common/utils/log"
)

const (
	backendChartServerAddr = "BACKEND_CHART_SERVER"
	namespaceParam         = ":repo"

	accessLevelPublic = iota
	accessLevelRead
	accessLevelWrite
	accessLevelAll
	accessLevelSystem
)

//chartController is a singleton instance
var chartController = initializeChartController()

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
	if incomingURI != "/index.yaml" && incomingURI != "/api/chartserver/health" {
		if !cra.requireNamespace(cra.namespace) {
			return
		}
	}
}

//GetHealthStatus handles GET /api/chartserver/health
func (cra *ChartRepositoryAPI) GetHealthStatus() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelSystem) {
		return
	}

	//Override the request path to '/health'
	req := cra.Ctx.Request
	req.URL.Path = "/health"

	chartController.GetBaseHandler().GetHealthStatus(cra.Ctx.ResponseWriter, req)
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
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelWrite) {
		return
	}

	chartController.GetManipulationHandler().UploadChartVersion(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

//UploadChartProvFile handles POST /api/:repo/prov
func (cra *ChartRepositoryAPI) UploadChartProvFile() {
	//Check access
	if !cra.requireAccess(cra.namespace, accessLevelWrite) {
		return
	}

	chartController.GetManipulationHandler().UploadProvenanceFile(cra.Ctx.ResponseWriter, cra.Ctx.Request)
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
		cra.RenderError(http.StatusInternalServerError, fmt.Sprintf("failed to check existence of namespace %s with error: %s", namespace, err.Error()))
		return false
	}

	//Not existing
	if !existsing {
		cra.HandleBadRequest(fmt.Sprintf("namespace %s is not existing", namespace))
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

	//At least, authentication is necessary when level > public
	if !cra.SecurityCtx.IsAuthenticated() {
		cra.HandleUnauthorized()
		return false
	}

	theLevel := accessLevel
	//If repo is empty, system admin role must be required
	if len(namespace) == 0 {
		theLevel = accessLevelSystem
	}

	switch theLevel {
	//Should be system admin role
	case accessLevelSystem:
		if !cra.SecurityCtx.IsSysAdmin() {
			cra.RenderError(http.StatusForbidden, fmt.Sprintf("system admin role is required but user '%s' is not", cra.SecurityCtx.GetUsername()))
			return false
		}
	case accessLevelAll:
		if !cra.SecurityCtx.HasAllPerm(namespace) {
			cra.RenderError(http.StatusForbidden, fmt.Sprintf("project admin role is required but user '%s' does not have", cra.SecurityCtx.GetUsername()))
			return false
		}
	case accessLevelWrite:
		if !cra.SecurityCtx.HasWritePerm(namespace) {
			cra.RenderError(http.StatusForbidden, fmt.Sprintf("developer role is required but user '%s' does not have", cra.SecurityCtx.GetUsername()))
			return false
		}
	case accessLevelRead:
		if !cra.SecurityCtx.HasReadPerm(namespace) {
			cra.RenderError(http.StatusForbidden, fmt.Sprintf("at least a guest role is required for user '%s'", cra.SecurityCtx.GetUsername()))
			return false
		}
	default:
		//access rejected for invalid scope
		cra.RenderError(http.StatusForbidden, "unrecognized access scope")
		return false
	}

	return true
}

//Initialize the chart service controller
func initializeChartController() *chartserver.Controller {
	addr := os.Getenv(backendChartServerAddr)
	url, err := url.Parse(addr)
	if err != nil {
		hlog.Fatal("chart storage server is not correctly configured")
	}

	controller, err := chartserver.NewController(url)
	if err != nil {
		hlog.Fatal("failed to initialize chart API controller")
	}

	hlog.Info("API controller for chart repository server is successfully initialized")

	return controller
}
