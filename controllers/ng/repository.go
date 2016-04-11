package ng

type RepositoryController struct {
	BaseController
}

func (rc *RepositoryController) Get() {
	rc.Forward("Repository", "repository.htm")
}
