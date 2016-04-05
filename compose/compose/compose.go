package compose

import (
	"gopkg.in/yaml.v2"
)

type SryCompose struct {
	CatalogConfig *CatalogConfig
	Applications  []*Application
	Graph         *ApplicationGraph
}

type CatalogConfig struct {
	Uuid              string    `json: "uuid" yaml: "uuid"`
	Name              string    `json: "name" yaml: "name"`
	Version           string    `json: "version" yaml: "version"`
	Description       string    `json: "description" yaml: "description"`
	MinimumSryVersion string    `json: "minimum_sry_version" yaml: "minimum_sry_version"`
	Questions         Questions `json: "questions" yaml: "questions"`
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

func (sc *SryCompose) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var params struct {
		CC CatalogConfig `json: "sry_catalog" yaml: "sry_catalog"`
	}

	if err := unmarshal(&params); err != nil {
		return err
	}

	sc.CatalogConfig = params.CC

	var apps map[string]*Application
	if err := unmarshal(&apps); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}

	//for k, v := range apps {
	//v.Name = k
	//sc.Applications = append(sc.Applications, v)
	//}

	return nil
}

func FromYaml(yamlString string) (*SryCompose, error) {
	compose := &SryCompose{}
	err := yaml.Unmarshal([]byte(yamlString), compose)
	if err != nil {
		return nil, err
	}

	return compose, nil
}

func FromJson(jsonString string) *SryCompose {
	compose := &SryCompose{}
	return compose
}

// SryCompose from omega-app
func FromInput(input interface{}) *SryCompose {
	return nil
}
