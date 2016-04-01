package compose

type Command func(ctx *Context) error

// create apps, POST to omega-app/apps for omega-app
func CommandCreate(ctx *Context) error {
	return nil
}

// return running status of an app as well as basic info
func CommandStatus(ctx *Context) error {
	return nil
}

// alias for create
func CommandUp(ctx *Context) error {
	return CommandCreate(ctx)
}

// stop app
func CommandStop(ctx *Context) error {
	return nil
}

// restarting app
func CommandRestart(ctx *Context) error {
	err := CommandStop(ctx)
	if err != nil {
		err = CommandUp(ctx)
	}
	return err
}
