package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/vmware/harbor/src/adminserver/client"
	"github.com/vmware/harbor/src/jobservice/config"
	"github.com/vmware/harbor/src/jobservice/env"
	"github.com/vmware/harbor/src/jobservice/job/impl"
	ilogger "github.com/vmware/harbor/src/jobservice/job/impl/logger"
	"github.com/vmware/harbor/src/jobservice/logger"
	"github.com/vmware/harbor/src/jobservice/runtime"
	"github.com/vmware/harbor/src/jobservice/utils"
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

	//Load configurations
	if err := config.DefaultConfig.Load(*configPath, true); err != nil {
		fmt.Printf("Failed to load configurations with error: %s\n", err)
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

	//New logger for job service
	sLogger := ilogger.NewServiceLogger(config.GetLogLevel())
	logger.SetLogger(sLogger)

	//Start
	runtime.JobService.LoadAndRun()
}
