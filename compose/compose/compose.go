package compose

type SryCompose struct {
	CatalogConfig      CatalogConfig `json: ".catalog" yaml: ".catalog"`
	ApplicationsConfig []Application
	Graph              ApplicationGraph
}

type CatalogConfig struct {
	Name              string `json: "name" yaml: "name"`
	Version           string `json: "version" yaml: "version"`
	Description       string `json: "description" yaml: "description"`
	Uuid              string `json: "uuid" yaml: "uuid"`
	MinimumSryVersion string `json: "minimum_sry_version" yaml: "minimum_sry_version"`
	Questions         Questions
}

type Question struct {
	Variable    string   `json: "variable" yaml: "variable"`
	Description string   `json: "description" yaml: "description"`
	Label       string   `json: "label" yaml: "label"`
	Type        string   `json: "type" yaml: "type"`
	Required    bool     `json: "required" yaml: "required"`
	Default     string   `json: "default" yaml: "default"`
	Options     []string `json: "options" yaml: "options"`
}

type Questions []Question

type Answer struct {
	Key   string
	Value string
}

func FromYaml(yamlString string) *SryCompose {
	compose := &SryCompose{}
	return compose
}

func FromJson(jsonString string) *SryCompose {
	compose := &SryCompose{}
	return compose
}

// SryCompose from omega-app
func FromInput(input interface{}) *SryCompose {
	return nil
}
