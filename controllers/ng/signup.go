package ng

// SignUpController handles requests to /ng/sign_up
type SignUpController struct {
	BaseController
}

// Get renders sign up page
func (suc *SignUpController) Get() {
	suc.Forward("Sign Up", "sign-up.htm")
}
