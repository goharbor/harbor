package ng

type ProjectController struct {
	BaseController
}

func (pc *ProjectController) Get() {
	pc.Forward("My Projects", "project.htm")
}
