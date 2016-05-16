package ng

import "os"

type RepositoryController struct {
	BaseController
}

func (rc *RepositoryController) Get() {
	rc.Data["HarborRegUrl"] = os.Getenv("HARBOR_REG_URL")
	rc.Forward("Repository", "repository.htm")
}
