package replication

type ImgOutParm struct {
	Secret  string          `json:"secret"`
	Image   string          `json:"image"`
	Targets []*RegistryInfo `json:"targets"`
}

type RegistryInfo struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}
