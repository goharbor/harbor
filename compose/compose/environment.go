package compose

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type Environment []Env
type Env struct {
	Key   string
	Value string
}

func (envs *Environment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var pair map[string]string
	if err := unmarshal(&pair); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}
	for k, v := range pair {
		*envs = append(*envs, Env{Key: k, Value: v})
	}
	return nil
}

func (e *Environment) ToString() string {
	envStr := ""
	for _, v := range *e {
		envStr += fmt.Sprintf(" Key: %s\n Value: %s\n", v.Key, v.Value)
	}

	return envStr
}
