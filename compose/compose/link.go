package compose

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"strings"
)

type Link struct {
	From   string
	Target string
}

func (l *Link) UnmarshalYAML(unmarshal func(interface{}) error) error {
	log.Println("links")
	var pair string
	if err := unmarshal(&pair); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}
	linksConfig := pair
	splited := strings.Split(linksConfig, ":")
	if len(splited) == 1 {
		l.From = splited[0]
		l.Target = splited[0]
	} else if len(splited) == 2 {
		l.From = splited[0]
		l.Target = splited[1]
	}
	return nil
}

func (l *Link) ToString() string {
	return fmt.Sprintf(" From: %-30s\n Target: %-30s\n ",
		l.From, l.Target)
}
