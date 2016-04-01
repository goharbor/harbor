package compose

type Context struct {
	Output  *interface{}
	Compose *SryCompose
}

// main entrance here, should be main when standalone mode
func EntryPoint() {}

func (ctx *Context) SetOutput(output *interface{}) {
	ctx.Output = output
}
