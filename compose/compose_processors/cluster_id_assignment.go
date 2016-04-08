package compose_processors

import (
	"github.com/vmware/harbor/compose/compose"
	"strconv"
)

func init() {
	Processors = append(Processors, ClusterIdAssignment)
}

func ClusterIdAssignment(sry_compose *compose.SryCompose) *compose.SryCompose {
	clusterId, ok := sry_compose.Answers["cluster_id"]
	if !ok {
		clusterId, ok = sry_compose.Answers["clusterid"]
	}

	for _, app := range sry_compose.Applications {
		clusterId_, _ := strconv.Atoi(clusterId)
		app.ClusterId = int32(clusterId_)
	}
	return sry_compose
}
