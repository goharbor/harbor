package apilib

//RoleParam ...
type RoleParam struct {
	Roles    []int32 `json:"roles,omitempty"`
	UserName string  `json:"user_name,omitempty"`
}
