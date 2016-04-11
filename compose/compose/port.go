package compose

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"strconv"
	"strings"
)

type Port struct {
	HostAddr      string
	HostPort      int
	ContainerAdd  string
	ContainerPort int
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
		hostPort, _ := strconv.Atoi(splited[0])
		containerPort, _ := strconv.Atoi(splited[0])
		p.HostPort = hostPort
		p.ContainerPort = containerPort
	} else if len(splited) == 2 {
		hostPort, _ := strconv.Atoi(splited[0])
		containerPort, _ := strconv.Atoi(splited[1])
		p.HostPort = hostPort
		p.ContainerPort = containerPort
	} else if len(splited) == 3 {
		p.HostAddr = splited[0]
		hostPort, _ := strconv.Atoi(splited[1])
		containerPort, _ := strconv.Atoi(splited[2])
		p.HostPort = hostPort
		p.ContainerPort = containerPort
	}
	return nil
}

func (p *Port) ToString() string {
	return fmt.Sprintf(" HostAddr: %-30s\n HostPort: %-30d\n ContainerAddr: %s\n ContainerPort: %d\n Protocol: %s\n",
		p.HostAddr, p.HostPort, p.ContainerAdd, p.ContainerPort, p.Protocol)
}
