package compose_processors

import (
	"github.com/vmware/harbor/compose/compose"
	"strconv"
)

func init() {
	Processors = append(Processors, OmegaAppCustomParameters)
}

func OmegaAppCustomParameters(sry_compose *compose.SryCompose) *compose.SryCompose {
	clusterId, ok := sry_compose.Answers["cluster_id"]
	if !ok {
		clusterId, ok = sry_compose.Answers["clusterid"]
	}

	for _, app := range sry_compose.Applications {
		clusterId_, _ := strconv.Atoi(clusterId)
		app.ClusterId = int32(clusterId_)
	}

	appName, ok := sry_compose.Answers["app_name"]
	if !ok {
		appName, ok = sry_compose.Answers["appname"]
	}

	for _, app := range sry_compose.Applications {
		app.AppName = appName
	}

	imageVersion, ok := sry_compose.Answers["image_version"]
	if !ok {
		imageVersion, ok = sry_compose.Answers["imageversion"]
	}

	for _, app := range sry_compose.Applications {
		app.ImageVersion = imageVersion
	}
	return sry_compose
}
