package ng

type SignUpController struct {
	BaseController
}

func (suc *SignUpController) Get() {
	suc.Forward("Sign Up", "sign-up.htm")
}
