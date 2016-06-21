package controllers

// DashboardController handles requests to /dashboard
type DashboardController struct {
	BaseController
}

// Get renders the dashboard  page
func (dc *DashboardController) Get() {
	dc.Forward("Dashboard", "dashboard.htm")
}
