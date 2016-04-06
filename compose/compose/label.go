package compose

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type Labels []Label
type Label struct {
	Key   string
	Value string
}

func (labels *Labels) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var pair map[string]string
	if err := unmarshal(&pair); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}
	for k, v := range pair {
		*labels = append(*labels, Label{Key: k, Value: v})
	}

	fmt.Println(labels)
	return nil
}

func (l *Labels) ToString() string {
	labelStr := ""
	for _, v := range *l {
		labelStr += fmt.Sprintf(" Key: %s\n Value: %s\n", v.Key, v.Value)
	}

	return labelStr
}
