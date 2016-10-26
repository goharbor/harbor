package controllers

import (
	"net/http"
)

// AccountSettingController handles request to /account_setting
type AccountSettingController struct {
	BaseController
}

// Get renders the account settings page
func (asc *AccountSettingController) Get() {
	if asc.AuthMode != "db_auth" {
		asc.CustomAbort(http.StatusForbidden, "")
	}
	asc.Forward("page_title_account_setting", "account-settings.htm")
}
