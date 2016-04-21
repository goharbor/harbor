package compose

import (
	"fmt"
	"strings"
)

const (
	DEFAULT_NET = "BRIDGE"
)

// compatiable with docker compose
type Application struct {
	MarathonConfig *MarathonConfig
	IsPrimary      bool // application depends on other applications
	MeetCritia     bool // application running now meet critia specified by compose file

	Name        string      `json:"name" yaml:"name"`
	Image       string      `json:"image" yaml:"image"`
	Command     interface{} `json:"command" yaml:"command"`
	EntryPoint  string      `json:"entrypoint" yaml:"entrypoint"`
	Environment Environment `json:"environment" yaml:"environment"`
	Labels      Labels      `json:"labels" yaml:"labels"`
	Volumes     []*Volume   `json:"volumes" yaml:"volumes"`
	Expose      []int       `json:"expose" yaml:"expose"`
	Ports       []*Port     `json:"ports" yaml:"ports"`
	Links       []*Link     `json:"links" yaml:"links"`
	Net         string      `json:"net" yaml:"net"`                   // bridge, host
	NetworkMode string      `json:"network_mode" yaml:"network_mode"` //compose version2 for net, same as net
	Restart     string      `json:"restart" yaml:"restart"`

	Dependencies []*Application
}

func (self *Application) Defaultlize() {
	if self.Net == "" {
		self.Net = DEFAULT_NET
	}

	if self.Command == nil {
		self.Command = interface{}("")
	}
}

func (app *Application) ToString() string {
	appBasic := ""
	appBasic = "\n"
	appBasic += app.MarathonConfig.ToString()
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

	appBasic += "Ports: \n\n"
	for _, v := range app.Ports {
		appBasic += fmt.Sprintf("%s\n", v.ToString())
	}

	appBasic += "Links: \n\n"
	for _, v := range app.Links {
		appBasic += fmt.Sprintf("%s\n", v.ToString())
	}
	return appBasic
}

func (app *Application) FormatedCommand() string {
	cmd := ""
	switch app.Command.(type) {
	case string:
		cmd = app.Command.(string)
	default:
		cmds := []string{}
		for _, v := range app.Command.([]interface{}) {
			cmds = append(cmds, v.(string))
		}
		cmd = strings.Join(cmds, " ")
	}
	return cmd
}
