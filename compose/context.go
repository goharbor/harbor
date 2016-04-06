package compose

import (
	"github.com/vmware/harbor/compose/command"
	"github.com/vmware/harbor/compose/compose"
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
func EntryPoint(yaml string, anwsers map[string]string, command command.Command) error {
	// default output/input channel now is OmegaAppOutput
	compose, err := compose.FromYaml(yaml)
	if err != nil {
		return err
	}

	ctx := &Context{
		Compose: compose,
		Command: command,
	}

	log.Println(compose.Catalog.ToString())
	log.Println()
	for _, v := range ctx.Compose.Applications {
		log.Println(v.ToString())
		log.Println()
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
	log.Println("putsxxxxxxxx ApplyChange")
	return nil
	//return ctx.Command(ctx.OutputChannel.(*channel.ChannelOutput), ctx.Compose)
}
