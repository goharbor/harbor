package models

// MemberQuery ...
type MemberQuery struct {
	UserID   int    // the user id
	Name     string // the username of member
	Role     int    // the role of the member has to the project
	GroupIDs []int  // the group ID of current user belongs to

	WithPublic bool // include the public projects for the member
}
