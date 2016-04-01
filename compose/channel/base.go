package channel

// for issue command to a channel
// eg. send app creation command to omega-app,
// k8s or swarm in future
type ChannelOutput interface {
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
