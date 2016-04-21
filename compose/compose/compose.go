package compose

import (
	"gopkg.in/yaml.v2"
)

type SryCompose struct {
	Catalog *Catalog
	Answers map[string]string

	Applications    []*Application
	MarathonConfigs []*MarathonConfig

	Graph *ApplicationGraph
}

func (sc *SryCompose) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// catalog
	var params struct {
		Catalog *Catalog `yaml: "catalog"`
	}
	if err := unmarshal(&params); err != nil {
		return err
	}
	sc.Catalog = params.Catalog

	// applications
	var apps map[string]*Application
	if err := unmarshal(&apps); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}
	for k, v := range apps {
		if k == "catalog" { // bypass yaml key catalog
			continue
		}

		v.Name = k
		v.Defaultlize()
		sc.Applications = append(sc.Applications, v)
	}

	return nil
}

func FromYaml(catalogContent, dockerComposeContent, marathonConfigContent string) (*SryCompose, error) {
	compose := &SryCompose{}

	catalog := new(Catalog)
	catalog, err := yaml.Unmarshal([]byte(catalogContent), catalog)
	if (err != nil) {
		log.Println("catalog file parse error")
		return nil, err
	}

	marathonConfig := new(MarathonConfig)
	marathonConfig, err := yaml.Unmarshal([]byte(marathonConfigContent), marathonConfig)


	compose.Catalog = 


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
