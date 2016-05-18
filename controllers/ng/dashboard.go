package ng

type DashboardController struct {
	BaseController
}

func (dc *DashboardController) Get() {
	dc.Forward("Dashboard", "dashboard.htm")
}
