package compose_processors

import (
	"github.com/vmware/harbor/compose/compose"
)

func init() {
	Processors = append(Processors, Graphnize)
}

func Graphnize(sry_compose *compose.SryCompose) *compose.SryCompose {
	graph := &compose.ApplicationGraph{}
	sry_compose.Graph = graph
	for _, app := range sry_compose.Applications {
		sry_compose.Graph.PrimaryApplications = append(sry_compose.Graph.PrimaryApplications, app)
	}

	return sry_compose
}
