package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/vmware/harbor/src/adminserver/client"
	"github.com/vmware/harbor/src/jobservice_v2/config"
	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/job/impl"
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

	//Set job context initializer
	runtime.JobService.SetJobContextInitializer(func(ctx *env.Context) (env.JobContext, error) {
		secret := config.GetAuthSecret()
		if utils.IsEmptyStr(secret) {
			return nil, errors.New("empty auth secret")
		}

		adminClient := client.NewClient(config.GetAdminServerEndpoint(), &client.Config{Secret: secret})
		jobCtx := impl.NewContext(ctx.SystemContext, adminClient)

		if err := jobCtx.Init(); err != nil {
			return nil, err
		}

		return jobCtx, nil
	})

	//Start
	runtime.JobService.LoadAndRun(*configPath, true)
}
