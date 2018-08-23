// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package gc

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/registryctl"
	reg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/registryctl/client"
)

const (
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = time.Minute + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
)

// GarbageCollector is the struct to run registry's garbage collection
type GarbageCollector struct {
	registryCtlClient client.Client
	logger            logger.Interface
	uiclient          *common_http.Client
	UIURL             string
	insecure          bool
	redisURL          string
}

// MaxFails implements the interface in job/Interface
func (gc *GarbageCollector) MaxFails() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (gc *GarbageCollector) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (gc *GarbageCollector) Validate(params map[string]interface{}) error {
	return nil
}

// Run implements the interface in job/Interface
func (gc *GarbageCollector) Run(ctx env.JobContext, params map[string]interface{}) error {
	if err := gc.init(ctx, params); err != nil {
		return err
	}
	if err := gc.readonly(true); err != nil {
		return err
	}
	defer gc.readonly(false)
	if err := gc.registryCtlClient.Health(); err != nil {
		gc.logger.Errorf("failed to start gc as regsitry controller is unreachable: %v", err)
		return err
	}
	gc.logger.Infof("start to run gc in job.")
	gcr, err := gc.registryCtlClient.StartGC()
	if err != nil {
		gc.logger.Errorf("failed to get gc result: %v", err)
		return err
	}
	if err := gc.cleanCache(); err != nil {
		return err
	}
	gc.logger.Infof("GC results: status: %t, message: %s, start: %s, end: %s.", gcr.Status, gcr.Msg, gcr.StartTime, gcr.EndTime)
	gc.logger.Infof("success to run gc in job.")
	return nil
}

func (gc *GarbageCollector) init(ctx env.JobContext, params map[string]interface{}) error {
	registryctl.Init()
	gc.registryCtlClient = registryctl.RegistryCtlClient
	gc.logger = ctx.GetLogger()
	cred := auth.NewSecretAuthorizer(os.Getenv("JOBSERVICE_SECRET"))
	gc.insecure = false
	gc.uiclient = common_http.NewClient(&http.Client{
		Transport: reg.GetHTTPTransport(gc.insecure),
	}, cred)
	errTpl := "Failed to get required property: %s"
	if v, ok := ctx.Get(common.UIURL); ok && len(v.(string)) > 0 {
		gc.UIURL = v.(string)
	} else {
		return fmt.Errorf(errTpl, common.UIURL)
	}
	gc.redisURL = params["redis_url_reg"].(string)
	return nil
}

func (gc *GarbageCollector) readonly(switcher bool) error {
	if err := gc.uiclient.Put(fmt.Sprintf("%s/api/configurations", gc.UIURL), struct {
		ReadOnly bool `json:"read_only"`
	}{
		ReadOnly: switcher,
	}); err != nil {
		gc.logger.Errorf("failed to send readonly request to %s: %v", gc.UIURL, err)
		return err
	}
	gc.logger.Info("the readonly request has been sent successfully")
	return nil
}

// cleanCache is to clean the registry cache for GC.
// To do this is because the issue https://github.com/docker/distribution/issues/2094
func (gc *GarbageCollector) cleanCache() error {

	con, err := redis.DialURL(
		gc.redisURL,
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)

	if err != nil {
		gc.logger.Errorf("failed to connect to redis %v", err)
		return err
	}
	defer con.Close()

	// clean all keys in registry redis DB.
	_, err = con.Do("FLUSHDB")
	if err != nil {
		gc.logger.Errorf("failed to clean registry cache %v", err)
		return err
	}

	return nil
}
