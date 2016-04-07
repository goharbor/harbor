package channel

import (
	"github.com/vmware/harbor/compose/command"
	"github.com/vmware/harbor/compose/compose"
)

// for issue command to a channel
// eg. send app creation command to omega-app,
// k8s or swarm in future
type ChannelOutput interface {
	Run(sry_compose *compose.SryCompose, cmd command.Command) error
	Create() error
	Stop() error
	Scale() error
	Restart() error
}

// for collect app status from a channel
// eg. currently omega-app supported
// k8s or swarm in future
type ChannelInput interface{}

type ComposeChannel interface {
	ChannelOutput
	ChannelInput
}
