package compose

type Context struct {
	OutputChannel *interface{}
	InputChannel  *interface{}
	Compose       *SryCompose
	Command       Command
}

// main entrance here, should be main when standalone mode
// should be **func main** in cli mode
func EntryPoint(yaml string, anwsers map[string]string, command Command) error {
	// default output/input channel now is OmegaAppOutput
	ctx := &Context{
		Compose: FromYaml(yaml),
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
	return ctx.Command(ctx)
}
