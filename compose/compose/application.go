package compose

import (
	"fmt"
	"github.com/vmware/harbor/utils"
	"gopkg.in/yaml.v2"
	"strings"
)

const (
	DEFAULT_CPU = 0.2
	DEFAULT_MEM = 2
	DEFAULT_NET = "bridge"
)

type Application struct {
	IsPrimary   bool          // application depends on other applications
	MeetCritia  bool          // application running now meet critia specified by compose file
	Name        string        `json: "name" yaml: "name"`
	Image       string        `json: "image" yaml: "image"`
	Cmd         string        `json: "cmd" yaml: "cmd"`
	EntryPoint  string        `json: "entrypoint" yaml: "entrypoint"`
	Cpu         float32       `json: "cpu" yaml: "cpu"`
	Mem         float32       `json: "mem" yaml: "mem"`
	Environment []Environment `json: "environment" yaml: "environment"`
	Labels      []*Label      `json: "labels" yaml: "labels"`
	Volume      []*Volume     `json: "volumes" yaml: "volumes"`
	Expose      []int         `json: "expose" yaml: "expose"`
	Port        []*Port       `json: "ports" yaml: "ports"`
	Net         string        `json: "net" yaml: "net"`
	Restart     string        `json: "restart" yaml: "restart"`

	Dependencies []*Application
}

func (self *Application) Defaultlize() {
	if utils.FloatEquals(self.Cpu, 0.0) {
		self.Cpu = DEFAULT_CPU
	}

	if utils.FloatEquals(self.Mem, 0.0) {
		self.Mem = DEFAULT_MEM
	}

	if self.Net == "" {
		self.Net = DEFAULT_NET
	}
}

type Label struct {
	Key   string
	Value string
}

type Volume struct {
	Container  string
	Host       string
	Permission string
}

func (app *Application) ToString() string {
	appBasic := ""
	appBasic = "\n"
	appBasic += fmt.Sprintf("Name: %-30s\n", app.Name)
	appBasic += fmt.Sprintf("Image: %-30s\n", app.Image)
	appBasic += fmt.Sprintf("Cmd: %-30s\n", app.Cmd)
	appBasic += fmt.Sprintf("EntryPoint: %-30s\n", app.EntryPoint)
	appBasic += fmt.Sprintf("Cpu: %-30f\n", app.Cpu)
	appBasic += fmt.Sprintf("Mem: %-30f\n", app.Mem)
	appBasic += fmt.Sprintf("Net: %-30s\n", app.Net)
	appBasic += fmt.Sprintf("Restart: %-30s\n", app.Restart)

	appBasic += "ENVS: \n\n"
	for _, v := range app.Environment {
		appBasic += fmt.Sprintf("%s\n", v.ToString())
	}

	appBasic += "Labels: \n\n"
	for _, v := range app.Labels {
		appBasic += fmt.Sprintf("%s\n", v.ToString())
	}

	appBasic += "Volumes: \n\n"
	for _, v := range app.Volume {
		appBasic += fmt.Sprintf("%s\n", v.ToString())
	}

	return appBasic
}

func (e *Label) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

func (l *Label) ToString() string {
	return fmt.Sprintf("Key %s\nValue %s\n", l.Key, l.Value)
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

type Port struct {
	HostAddr      string
	HostPort      int
	ContaienrAddr string
	ContaienrPort int
	Protocol      string
}
