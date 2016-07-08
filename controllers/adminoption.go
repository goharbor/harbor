package controllers

// AdminOptionController handles requests to /admin_option
type AdminOptionController struct {
	BaseController
}

// Get renders the admin options  page
func (aoc *AdminOptionController) Get() {
	aoc.Forward("page_title_admin_option", "admin-options.htm")
}
