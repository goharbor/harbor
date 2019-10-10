package jfrog

type repository struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	PackageType string `json:"packageType"`
}

type repositoryCreate struct {
	Key           string `json:"key"`
	Rclass        string `json:"rclass"`
	PackageType   string `json:"packageType"`
	RepoLayoutRef string `json:"repoLayoutRef"`
}

func newDefaultDockerLocalRepository(key string) *repositoryCreate {
	return &repositoryCreate{
		Key:           key,
		Rclass:        "local",
		PackageType:   "docker",
		RepoLayoutRef: "simple-default",
	}
}
