package compose

import (
	"github.com/vmware/harbor/compose/channel"
	"github.com/vmware/harbor/compose/command"
	"github.com/vmware/harbor/compose/compose"
)

type Context struct {
	OutputChannel interface{}
	InputChannel  interface{}
	Compose       *compose.SryCompose
	Command       command.Command
}

// main entrance here, should be main when standalone mode
// should be **func main** in cli mode
func EntryPoint(yaml string, anwsers map[string]string, command command.Command) error {
	// default output/input channel now is OmegaAppOutput
	ctx := &Context{
		Compose: compose.FromYaml(yaml),
		Command: command,
	}

	return ctx.ApplyChange()
}

func (ctx *Context) SetOutput(output *interface{}) {
	//ctx.Output = output
}

func (ctx *Context) SetInput(output *interface{}) {
	//ctx.Output = output
}

func (ctx *Context) ApplyChange() error {
	return ctx.Command(ctx.OutputChannel.(*channel.ChannelOutput), ctx.Compose)
}
