package controllers

// AccountSettingController handles request to /account_setting
type AccountSettingController struct {
	BaseController
}

// Get renders the account settings page
func (asc *AccountSettingController) Get() {
	var isAdminForLdap bool
	sessionUserID, ok := asc.GetSession("userId").(int)
	if !ok {
		asc.Redirect("/", 302)
	}
	if ok && sessionUserID == 1 {
		isAdminForLdap = true
	}
	if asc.AuthMode == "db_auth" || isAdminForLdap {
		asc.Forward("page_title_account_setting", "account-settings.htm")
	} else {
		asc.Redirect("/dashboard", 302)
	}
}
