package ng

// AccountSettingController handles request to /ng/account_setting
type AccountSettingController struct {
	BaseController
}

// Get renders the account settings page
func (asc *AccountSettingController) Get() {
	asc.Forward("Account Settings", "account-settings.htm")
}
