package compose_processors

import (
	"github.com/vmware/harbor/compose/compose"
)

type ComposeProcessor func(compose *compose.SryCompose) *compose.SryCompose

var Processors []ComposeProcessor
