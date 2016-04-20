package compose

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"strings"
)

type Port struct {
	HostAddr      string
	HostPort      string
	ContainerAddr string
	ContainerPort string
	Protocol      string
}

func (p *Port) UnmarshalYAML(unmarshal func(interface{}) error) error {
	log.Println("ports")
	var pair string
	if err := unmarshal(&pair); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}
	portsConfig := pair
	if strings.Contains(pair, "/") {
		portsConfig = strings.Split(pair, "/")[0]
		p.Protocol = strings.Split(pair, "/")[1]
	}

	splited := strings.Split(portsConfig, ":")
	if len(splited) == 1 {
		p.HostPort = splited[0]
		p.ContainerPort = splited[0]
	} else if len(splited) == 2 {
		p.HostPort = splited[0]
		p.ContainerPort = splited[1]
	} else if len(splited) == 3 {
		p.HostAddr = splited[0]
		p.HostPort = splited[1]
		p.ContainerPort = splited[2]
	}
	return nil
}

func (p *Port) ToString() string {
	return fmt.Sprintf(" HostAddr: %-30s\n HostPort: %-30s\n ContainerAddr: %s\n ContainerPort: %s\n Protocol: %s\n",
		p.HostAddr, p.HostPort, p.ContainerAddr, p.ContainerPort, p.Protocol)
}
