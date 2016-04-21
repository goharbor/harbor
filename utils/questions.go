package utils

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type Catalog struct {
	Uuid              string    `json:"uuid" yaml:"uuid"`
	Name              string    `json:"name" yaml:"name"`
	Version           string    `json:"version" yaml:"version"`
	Description       string    `json:"description" yaml:"description"`
	MinimumSryVersion string    `json:"minimum_sry_version" yaml:"minimum_sry_version"`
	Questions         Questions `json:"questions" yaml:"questions"`
}

type Question struct {
	Variable    string   `json:"variable" yaml:"variable"`
	Description string   `json:"description" yaml:"description"`
	Label       string   `json:"label" yaml:"label"`
	Type        string   `json:"type" yaml:"type"`
	Required    bool     `json:"required" yaml:"required"`
	Default     string   `json:"default" yaml:"default"`
	Options     []string `json:"options" yaml:"options"`
}

type Questions []Question

func (c *Catalog) ToString() string {
	catalog := ""
	catalog += fmt.Sprintf("\n")
	catalog += fmt.Sprintf("Name: %-30s\n", c.Name)
	catalog += fmt.Sprintf("Uuid: %-30s\n", c.Uuid)
	catalog += fmt.Sprintf("Version: %-30s\n", c.Version)
	catalog += fmt.Sprintf("Description: %-30s\n", c.Description)
	catalog += fmt.Sprintf("MinimumSryVersion: %-30s\n", c.MinimumSryVersion)

	return catalog
}

func ParseQuestions(catalogContent string) (*Catalog, error) {
	catalog := new(Catalog)
	err := yaml.Unmarshal([]byte(catalogContent), catalog)
	if err != nil {
		return nil, err
	}

	return catalog, nil
}
