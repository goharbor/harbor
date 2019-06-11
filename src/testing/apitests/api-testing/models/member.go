package models

// Member : For /api/projects/:pid/members
type Member struct {
	RoleID int         `json:"role_id"`
	Member *MemberUser `json:"member_user"`
}

// MemberUser ...
type MemberUser struct {
	Username string `json:"username"`
}

// ExistingMember : For GET /api/projects/20/members
type ExistingMember struct {
	MID    int    `json:"id"`
	Name   string `json:"entity_name"`
	RoleID int    `json:"role_id"`
}
