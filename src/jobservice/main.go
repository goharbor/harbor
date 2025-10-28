// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/job/impl"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/runtime"
	"github.com/goharbor/harbor/src/lib"
	cfgLib "github.com/goharbor/harbor/src/lib/config"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/notation"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/nydus"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/subject"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	_ "github.com/goharbor/harbor/src/pkg/config/rest"
	_ "github.com/goharbor/harbor/src/pkg/scan/sbom"
	_ "github.com/goharbor/harbor/src/pkg/scan/vulnerability"
)

func main() {
	// Start pprof server
	lib.StartPprof()

	cfgLib.DefaultCfgManager = common.RestCfgManager
	if err := cfgLib.DefaultMgr().Load(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to load configuration, error: %v", err))
	}

	// Get parameters
	configPath := flag.String("c", "", "Specify the yaml config file path")
	flag.Parse()

	// Missing config file
	if configPath == nil || utils.IsEmptyStr(*configPath) {
		flag.Usage()
		panic("no config file is specified")
	}

	// Load configurations
	if err := config.DefaultConfig.Load(*configPath, true); err != nil {
		panic(fmt.Sprintf("load configurations error: %s\n", err))
	}

	// Append node ID
	vCtx := context.WithValue(context.Background(), utils.NodeID, utils.GenerateNodeID())
	// Create the root context
	ctx, cancel := context.WithCancel(vCtx)
	defer cancel()

	// Initialize logger
	if err := logger.Init(ctx); err != nil {
		panic(err)
	}

	cfgLib.InitTraceConfig(ctx)
	defer tracelib.InitGlobalTracer(context.Background()).Shutdown()

	// Set job context initializer
	runtime.JobService.SetJobContextInitializer(func(ctx context.Context) (job.Context, error) {
		secret := config.GetAuthSecret()
		if utils.IsEmptyStr(secret) {
			return nil, errors.New("empty auth secret")
		}

		jobCtx := impl.NewContext(ctx, cfgLib.DefaultMgr())

		if err := jobCtx.Init(); err != nil {
			return nil, err
		}

		return jobCtx, nil
	})

	// Start
	if err := runtime.JobService.LoadAndRun(ctx, cancel); err != nil {
		logger.Fatal(err)
	}
}
