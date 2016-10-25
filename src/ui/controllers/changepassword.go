package controllers

import (
	"net/http"
)

// ChangePasswordController handles request to /change_password
type ChangePasswordController struct {
	BaseController
}

// Get renders the change password page
func (asc *ChangePasswordController) Get() {
	if asc.AuthMode != "db_auth" {
		asc.CustomAbort(http.StatusForbidden, "")
	}
	asc.Forward("page_title_change_password", "change-password.htm")
}
