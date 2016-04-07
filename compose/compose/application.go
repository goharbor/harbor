package compose

import (
	"fmt"
	"github.com/vmware/harbor/utils"
	"strings"
)

const (
	DEFAULT_CPU       = 0.2
	DEFAULT_MEM       = 2
	DEFAULT_INSTANCES = 2
	DEFAULT_NET       = "bridge"
)

type Application struct {
	IsPrimary   bool        // application depends on other applications
	MeetCritia  bool        // application running now meet critia specified by compose file
	Name        string      `json: "name" yaml: "name"`
	Image       string      `json: "image" yaml: "image"`
	Command     interface{} `json: "command" yaml: "command"`
	EntryPoint  string      `json: "entrypoint" yaml: "entrypoint"`
	Cpu         float32     `json: "cpu" yaml: "cpu"`
	Mem         float32     `json: "mem" yaml: "mem"`
	Instances   int32       `json: "instances" yaml: "instances"`
	Environment Environment `json: "environment" yaml: "environment"`
	Labels      Labels      `json: "labels" yaml: "labels"`
	Volumes     []*Volume   `json: "volumes" yaml: "volumes"`
	Expose      []int       `json: "expose" yaml: "expose"`
	Port        []*Port     `json: "ports" yaml: "ports"`
	Net         string      `json: "net" yaml: "net"`
	Restart     string      `json: "restart" yaml: "restart"`
	LogPaths    []string    `json: "log_paths" yaml: "log_paths"`

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

	if self.Instances == 0 {
		self.Instances = DEFAULT_INSTANCES
	}
}

func (app *Application) ToString() string {
	appBasic := ""
	appBasic = "\n"
	appBasic += fmt.Sprintf("Name: %-30s\n", app.Name)
	appBasic += fmt.Sprintf("Image: %-30s\n", app.Image)
	switch app.Command.(type) {
	case string:
		appBasic += fmt.Sprintf("Command: %-30s\n", app.Command.(string))
	default:
		cmds := []string{}
		for _, v := range app.Command.([]interface{}) {
			cmds = append(cmds, v.(string))
		}
		appBasic += fmt.Sprintf("Command: %-30s\n", strings.Join(cmds, " "))
	}
	appBasic += fmt.Sprintf("EntryPoint: %-30s\n", app.EntryPoint)
	appBasic += fmt.Sprintf("Cpu: %-30f\n", app.Cpu)
	appBasic += fmt.Sprintf("Mem: %-30f\n", app.Mem)
	appBasic += fmt.Sprintf("Instances: %-30d\n", app.Instances)
	appBasic += fmt.Sprintf("Net: %-30s\n", app.Net)
	appBasic += fmt.Sprintf("Restart: %-30s\n", app.Restart)

	appBasic += "ENVS: \n\n"
	appBasic += app.Environment.ToString()

	appBasic += "Labels: \n\n"
	appBasic += app.Labels.ToString()

	appBasic += "Volumes: \n\n"
	for _, v := range app.Volumes {
		appBasic += fmt.Sprintf("%s\n", v.ToString())
	}

	return appBasic
}
