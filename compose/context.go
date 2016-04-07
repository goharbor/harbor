package compose

import (
	"errors"
	"github.com/vmware/harbor/compose/channel"
	"github.com/vmware/harbor/compose/command"
	"github.com/vmware/harbor/compose/compose"
	"github.com/vmware/harbor/compose/compose_processors"
	"log"
)

type Context struct {
	OutputChannel interface{}
	InputChannel  interface{}
	Compose       *compose.SryCompose
	Command       command.Command
}

// main entrance here, should be main when standalone mode
// should be **func main** in cli mode
func EntryPoint(yaml string, anwsers map[string]string, command command.Command, config channel.ChannelHttpConfig) error {
	// default output/input channel now is OmegaAppOutput
	compose, err := compose.FromYaml(yaml)
	if err != nil {
		return err
	}
	compose.Answers = anwsers

	ctx := &Context{
		Compose:       compose,
		Command:       command,
		OutputChannel: channel.NewOmegaOutput(config),
	}

	for _, v := range compose_processors.Processors {
		compose = v(compose)
	}

	for _, v := range ctx.Compose.Applications {
		log.Println(v.ToString())
	}

	log.Println(len(ctx.Compose.Graph.PrimaryApplications))
	return ctx.ApplyChange()
}

func (ctx *Context) SetOutput(output *interface{}) {
	//ctx.Output = output
}

func (ctx *Context) SetInput(output *interface{}) {
	//ctx.Output = output
}

func (ctx *Context) ApplyChange() error {
	_, ok := ctx.OutputChannel.(channel.ChannelOutput)
	if ok {
		return ctx.OutputChannel.(channel.ChannelOutput).Run(ctx.Compose, ctx.Command)
	}
	return errors.New("failed output channel type assertion")
}
