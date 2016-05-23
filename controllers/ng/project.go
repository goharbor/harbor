package ng

// ProjectController handles requests to /ng/projec
type ProjectController struct {
	BaseController
}

// Get renders project page
func (pc *ProjectController) Get() {
	pc.Forward("My Projects", "project.htm")
}
