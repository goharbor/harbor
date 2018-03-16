package main

import (
	"flag"
	"fmt"

	"github.com/vmware/harbor/src/jobservice_v2/runtime"
	"github.com/vmware/harbor/src/jobservice_v2/utils"
)

func main() {
	//Get parameters
	configPath := flag.String("c", "", "Specify the yaml config file path")
	flag.Parse()

	//Missing config file
	if configPath == nil || utils.IsEmptyStr(*configPath) {
		fmt.Println("Config file should be specified")
		flag.Usage()
		return
	}

	//Start
	runtime.JobService.LoadAndRun(*configPath, true)
}
