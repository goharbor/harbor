package ng

type AccountSettingController struct {
	BaseController
}

func (asc *AccountSettingController) Get() {
	asc.Forward("Account Settings", "account-settings.htm")
}
