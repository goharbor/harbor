package command

import (
	"github.com/vmware/harbor/compose/channel"
	"github.com/vmware/harbor/compose/compose"
)

type Command func(output *channel.ChannelOutput, compose *compose.SryCompose) error

// create apps, POST to omega-app/apps for omega-app
func CommandCreate(output *channel.ChannelOutput, compose *compose.SryCompose) error {
	return nil
}

// return running status of an app as well as basic info
func CommandStatus(output *channel.ChannelOutput, compose *compose.SryCompose) error {
	return nil
}

// alias for create
func CommandUp(output *channel.ChannelOutput, compose *compose.SryCompose) error {
	return CommandCreate(output, compose)
}

// stop app
func CommandStop(output *channel.ChannelOutput, compose *compose.SryCompose) error {
	return nil
}

// restarting app
func CommandRestart(output *channel.ChannelOutput, compose *compose.SryCompose) error {
	err := CommandStop(output, compose)
	if err != nil {
		err = CommandUp(output, compose)
	}
	return err
}
