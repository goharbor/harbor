package controllers

// ProjectController handles requests to /project
type ProjectController struct {
	BaseController
}

// Get renders project page
func (pc *ProjectController) Get() {
	pc.Forward("page_title_project", "project.htm")
}
