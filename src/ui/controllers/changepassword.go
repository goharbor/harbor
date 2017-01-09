package controllers

// ChangePasswordController handles request to /change_password
type ChangePasswordController struct {
	BaseController
}

// Get renders the change password page
func (cpc *ChangePasswordController) Get() {
	var isAdminForLdap bool
	sessionUserID, ok := cpc.GetSession("userId").(int)
	if !ok {
		cpc.Redirect("/", 302)
	}
	if ok && sessionUserID == 1 {
		isAdminForLdap = true
	}
	if cpc.AuthMode == "db_auth" || isAdminForLdap {
		cpc.Forward("page_title_change_password", "change-password.htm")
	} else {
		cpc.Redirect("/dashboard", 302)
	}
}
