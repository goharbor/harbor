package gitlab

type TokenResp struct {
	Token string `json:"token"`
}

type Project struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	FullPath        string `json:"path_with_namespace"`
	Visibility      string `json:"visibility"`
	RegistryEnabled bool   `json:"container_registry_enabled"`
}

type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Location string `json:"location"`
}

type Tag struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Location string `json:"location"`
}
