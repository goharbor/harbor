package controllers

import (
	"net/http"
)

// SignUpController handles requests to /sign_up
type SignUpController struct {
	BaseController
}

// Get renders sign up page
func (suc *SignUpController) Get() {
	if suc.AuthMode != "db_auth" {
		suc.CustomAbort(http.StatusUnauthorized, "Status unauthorized.")
	}
	suc.Data["AddNew"] = false
	suc.Forward("Sign Up", "sign-up.htm")
}
