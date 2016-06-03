package HarborAPI

type User struct {
	UserId   int32  `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Realname string `json:"realname,omitempty"`
	Comment  string `json:"comment,omitempty"`
	Deleted  int32  `json:"deleted,omitempty"`
}
