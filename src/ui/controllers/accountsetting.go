package controllers

// AccountSettingController handles request to /account_setting
type AccountSettingController struct {
	BaseController
}

// Get renders the account settings page
func (asc *AccountSettingController) Get() {
	asc.Forward("page_title_account_setting", "account-settings.htm")
}
