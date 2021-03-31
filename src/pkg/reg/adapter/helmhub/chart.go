package helmhub

type chart struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type chartList struct {
	Data []*chart `json:"data"`
}

type chartAttributes struct {
	Version string   `json:"version"`
	URLs    []string `json:"urls"`
}

type chartRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type chartData struct {
	Name string     `json:"name"`
	Repo *chartRepo `json:"repo"`
}

type chartInfo struct {
	Data *chartData `json:"data"`
}

type chartRelationships struct {
	Chart *chartInfo `json:"chart"`
}

type chartVersion struct {
	ID            string              `json:"id"`
	Type          string              `json:"type"`
	Attributes    *chartAttributes    `json:"attributes"`
	Relationships *chartRelationships `json:"relationships"`
}

type chartVersionList struct {
	Data []*chartVersion `json:"data"`
}
