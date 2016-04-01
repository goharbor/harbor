package compose

type Command func(ctx *Context) error

func Create(ctx *Context) error {
	return nil
}

func Status(ctx *Context) error {
	return nil
}

func CommandFromString(command string) Command {
	switch command {
	case "create":
		return Create
	case "status":
		return Status
	}
	return nil
}
