package controllers

// ChangePasswordController handles request to /change_password
type ChangePasswordController struct {
	BaseController
}

// Get renders the change password page
func (asc *ChangePasswordController) Get() {
	asc.Forward("page_title_change_password", "change-password.htm")
}
