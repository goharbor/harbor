package controllers

import (
	"github.com/vmware/harbor/src/ui/config"
)

// RepositoryController handles request to /repository
type RepositoryController struct {
	BaseController
}

// Get renders repository page
func (rc *RepositoryController) Get() {
	rc.Data["HarborRegUrl"] = config.ExtRegistryURL()
	rc.Forward("page_title_repository", "repository.htm")
}
