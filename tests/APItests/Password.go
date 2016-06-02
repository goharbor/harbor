package HarborAPI

type Password struct {
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}
