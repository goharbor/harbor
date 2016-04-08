package compose

import (
	"errors"
	"log"
	"strings"

	"github.com/vmware/harbor/compose/channel"
	"github.com/vmware/harbor/compose/command"
	"github.com/vmware/harbor/compose/compose"
	"github.com/vmware/harbor/compose/compose_processors"
)

type Context struct {
	OutputChannel interface{}
	InputChannel  interface{}
	Compose       *compose.SryCompose
	Command       command.Command
}

// main entrance here, should be main when standalone mode
// should be **func main** in cli mode
func EntryPoint(yaml string, answers map[string]string, command command.Command, config channel.ChannelHttpConfig) error {
	// default output/input channel now is OmegaAppOutput
	compose, err := compose.FromYaml(yaml)
	if err != nil {
		return err
	}

	lowerCaseAnswers := make(map[string]string)
	for k, v := range answers {
		lowerCaseAnswers[strings.ToLower(k)] = strings.ToLower(v)
	}
	compose.Answers = lowerCaseAnswers

	log.Println(compose.Applications[0].Environment)

	ctx := &Context{
		Compose:       compose,
		Command:       command,
		OutputChannel: channel.NewOmegaOutput(config),
	}
	log.Println(compose.Applications[0].Environment)

	for _, v := range compose_processors.Processors {
		compose = v(compose)
	}

	for _, v := range ctx.Compose.Applications {
		log.Println(v.ToString())
	}

	return ctx.ApplyChange()
}

func ComposeParse(yaml string) (*compose.SryCompose, error) {
	compose, err := compose.FromYaml(yaml)
	if err != nil {
		return nil, err
	}
	return compose, nil
}

func (ctx *Context) SetOutput(output *channel.ChannelOutput) {
	ctx.OutputChannel = output
}

func (ctx *Context) SetInput(output *interface{}) {
	//ctx.Output = output
}

func (ctx *Context) ApplyChange() error {
	_, ok := ctx.OutputChannel.(*channel.OmegaAppOutput)
	log.Println(ok)
	if ok {
		return ctx.OutputChannel.(*channel.OmegaAppOutput).Run(ctx.Compose, ctx.Command)
	}
	return errors.New("failed output channel type assertion")
}
