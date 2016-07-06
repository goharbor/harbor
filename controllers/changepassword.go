package controllers

// AccountPwdSettingController handles request to /account_setting
type ChangePasswordController struct {
	BaseController
}

// Get renders the account settings page
func (asc *ChangePasswordController) Get() {
	asc.Forward("page_title_reset_password", "change-password.htm")
}
