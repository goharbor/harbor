package compose

type Command func(ctx *Context) error

func CommandCreate(ctx *Context) error {
	return nil
}

func CommandStatus(ctx *Context) error {
	return nil
}
