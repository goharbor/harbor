package compose

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

type Environment struct {
	Key   string
	Value string
}

func (e *Environment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var pair string
	if err := unmarshal(&pair); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}
	if len(pair) > 0 {
		e.Key = strings.Split(pair, ":")[0]
		e.Value = strings.Split(pair, ":")[1]
	}
	return nil
}

func (e *Environment) ToString() string {
	return fmt.Sprintf("Key %s\nValue %s\n", e.Key, e.Value)
}
