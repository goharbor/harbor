package compose_processors

import (
	"github.com/vmware/harbor/compose/compose"
)

func init() {
	Processors = append(Processors, UnitStandarization)
}

func UnitStandarization(compose *compose.SryCompose) *compose.SryCompose {
	return compose
}
