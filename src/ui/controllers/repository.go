package controllers

import (
	"net/http"
	"strings"

	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

// RepositoryController handles request to /repository
type RepositoryController struct {
	BaseController
}

// Get renders repository page
func (rc *RepositoryController) Get() {
	url, err := config.ExtEndpoint()
	if err != nil {
		log.Errorf("failed to get domain name: %v", err)
		rc.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	rc.Data["HarborRegUrl"] = strings.Split(url, "://")[1]
	rc.Forward("page_title_repository", "repository.htm")
}
