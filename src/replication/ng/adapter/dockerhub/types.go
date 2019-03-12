package dockerhub

// LoginCredential is request to login.
type LoginCredential struct {
	User     string `json:"username"`
	Password string `json:"password"`
}

// TokenResp is response of login.
type TokenResp struct {
	Token string `json:"token"`
}

// NamespacesResp is namespace list responsed from DockerHub.
type NamespacesResp struct {
	// Namespaces is a list of namespaces
	Namespaces []string `json:"namespaces"`
}

// NewOrgReq is request to create a new org as namespace.
type NewOrgReq struct {
	// Name is name of the namespace
	Name string `json:"orgname"`
	// FullName ...
	FullName string `json:"full_name"`
	// Company ...
	Company string `json:"company"`
	// Location ...
	Location string `json:"location"`
	// ProfileUrl ...
	ProfileURL string `json:"profile_url"`
	// GravatarEmail ...
	GravatarEmail string `json:"gravatar_email"`
}
