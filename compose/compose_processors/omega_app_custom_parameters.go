package compose_processors

import (
	"github.com/vmware/harbor/compose/compose"
	"log"
	"strconv"
)

func init() {
	Processors = append(Processors, OmegaAppCustomParameters)
}

func OmegaAppCustomParameters(sry_compose *compose.SryCompose) *compose.SryCompose {
	// cluster id
	clusterId, ok := sry_compose.Answers["cluster_id"]
	if !ok {
		clusterId, ok = sry_compose.Answers["clusterid"]
	}

	for _, app := range sry_compose.Applications {
		clusterId_, _ := strconv.ParseFloat(clusterId, 32)
		log.Println(clusterId)
		app.ClusterId = int32(clusterId_)
		log.Println("cluster id")
		log.Println(app.ClusterId)
	}

	// appname
	appName, ok := sry_compose.Answers["app_name"]
	if !ok {
		appName, ok = sry_compose.Answers["appname"]
	}

	for _, app := range sry_compose.Applications {
		app.AppName = appName
	}

	// image version
	imageVersion, ok := sry_compose.Answers["image_version"]
	if !ok {
		imageVersion, ok = sry_compose.Answers["imageversion"]
	}

	for _, app := range sry_compose.Applications {
		app.ImageVersion = imageVersion
	}

	// cpu
	cpu, ok := sry_compose.Answers["cpus"]
	if !ok {
		cpu, ok = sry_compose.Answers["cpus"]
	}

	for _, app := range sry_compose.Applications {
		cpu_, _ := strconv.ParseFloat(cpu, 32)
		app.Cpu = float32(cpu_)
	}

	// mem
	mem, ok := sry_compose.Answers["mem"]
	if !ok {
		mem, ok = sry_compose.Answers["mem"]
	}

	for _, app := range sry_compose.Applications {
		mem_, _ := strconv.ParseFloat(mem, 32)
		app.Mem = float32(mem_)
	}

	// instances
	instances, ok := sry_compose.Answers["instances"]
	if !ok {
		instances, ok = sry_compose.Answers["instances"]
	}

	for _, app := range sry_compose.Applications {
		instances_, _ := strconv.ParseFloat(instances, 32)
		app.Instances = int32(instances_)
	}

	return sry_compose
}
