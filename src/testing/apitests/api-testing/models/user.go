package models

// User : For /api/users
type User struct {
	Username string `json:"username"`
	RealName string `json:"realname"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Comment  string `json:"comment"`
}

// ExistingUser : For GET /api/users
type ExistingUser struct {
	User
	ID int `json:"user_id"`
}
