package compose_processors

import (
	"github.com/vmware/harbor/compose/compose"
)

func init() {
	Processors = append(Processors, UnitStandarization)
}

// eg. turn MB to 1024 * 1024
func UnitStandarization(compose *compose.SryCompose) *compose.SryCompose {
	return compose
}
