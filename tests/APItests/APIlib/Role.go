package HarborAPI

type Role struct {
	RoleId   int32  `json:"role_id,omitempty"`
	RoleCode string `json:"role_code,omitempty"`
	RoleName string `json:"role_name,omitempty"`
}
