package compose

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

type Application struct {
	IsPrimary   bool           // application depends on other applications
	MeetCritia  bool           // application running now meet critia specified by compose file
	Name        string         `json: "name" yaml: "name"`
	Image       string         `json: "image" yaml: "image"`
	Cmd         string         `json: "cmd" yaml: "cmd"`
	EntryPoint  string         `json: "entrypoint" yaml: "entrypoint"`
	Cpu         float32        `json: "cpu" yaml: "cpu"`
	Mem         float32        `json: "mem" yaml: "mem"`
	Environment []*Environment `json: "environment" yaml: "environment"`
	Labels      []*Label       `json: "labels" yaml: "labels"`
	Volume      []*Volume      `json: "volumes" yaml: "volumes"`
	Expose      []int          `json: "expose" yaml: "expose"`
	Port        []*Port        `json: "ports" yaml: "ports"`
	Net         string         `json: "net" yaml: "net"`
	Restart     string         `json: "restart" yaml: "restart"`

	Dependencies []*Application
}

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

type Label struct {
	Key   string
	Value string
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

func (app *Application) ToString() string {
	appBasic := fmt.Sprintf("Column   Value\nName  %s\nImage %s\nCmd %s\nEntryPoint %s\nCpu %f\nMem  %f\nNet %s\nRestart %s\n",
		app.Name, app.Image, app.Cmd, app.EntryPoint, app.Cpu, app.Mem, app.Net, app.Restart)
	fmt.Println(len(app.Environment))
	fmt.Println(len(app.Labels))
	fmt.Println(len(app.Volume))

	for _, v := range app.Environment {
		appBasic += v.ToString()
	}

	for _, v := range app.Labels {
		appBasic += v.ToString()
	}

	for _, v := range app.Volume {
		appBasic += v.ToString()
	}

	return appBasic
}

type Port struct {
	HostAddr      string
	HostPort      int
	ContaienrAddr string
	ContaienrPort int
	Protocol      string
}
