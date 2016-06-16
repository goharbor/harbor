package controllers

// AccountSettingController handles request to /account_setting
type AccountSettingController struct {
	BaseController
}

// Get renders the account settings page
func (asc *AccountSettingController) Get() {
	asc.Forward("Account Settings", "account-settings.htm")
}
