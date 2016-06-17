package controllers

import "os"

// RepositoryController handles request to /repository
type RepositoryController struct {
	BaseController
}

// Get renders repository page
func (rc *RepositoryController) Get() {
	rc.Data["HarborRegUrl"] = os.Getenv("HARBOR_REG_URL")
	rc.Forward("Repository", "repository.htm")
}
