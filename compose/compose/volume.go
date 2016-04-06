package compose

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

type Volume struct {
	Container  string
	Host       string
	Permission string
}

func (v *Volume) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var pair string
	if err := unmarshal(&pair); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}
	splited := strings.Split(pair, ":")[0]
	if len(splited) == 1 {
		v.Container = string(splited[0])
		v.Host = string(splited[0])
		v.Permission = "rw"
	} else if len(splited) == 2 {
		v.Container = string(splited[0])
		v.Host = string(splited[1])
		v.Permission = "rw"
	} else if len(splited) == 3 {
		v.Container = string(splited[0])
		v.Host = string(splited[1])
		v.Permission = strings.ToLower(string(splited[2]))
	}
	return nil
}

func (v *Volume) ToString() string {
	return fmt.Sprintf("Host %s\nContainer %s\n Permission %s\n", v.Host, v.Container, v.Permission)
}
