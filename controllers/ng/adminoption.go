package ng

// AdminOptionController handles requests to /ng/admin_option
type AdminOptionController struct {
	BaseController
}

// Get renders the admin options  page
func (aoc *AdminOptionController) Get() {
	aoc.Forward("Admin Options", "admin-options.htm")
}
