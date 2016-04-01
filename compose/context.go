package compose

type Context struct {
	Output  *interface{}
	Compose *SryCompose
	Command *Command
}

// main entrance here, should be main when standalone mode
func EntryPoint(yaml string, anwsers map[string]string, command Command) error {
	ctx := &Context{
		Compose: compose.FromYaml(yaml),
		Command: command,
	}

	return ctx.Run()
}

func (ctx *Context) SetOutput(output *interface{}) {
	ctx.Output = output
}

func (ctx *Context) Run() error {
	ctx.Command(ctx)
	return nil
}
