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

package systeminfo

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/systeminfo"
	"github.com/goharbor/harbor/src/pkg/systeminfo/imagestorage"
	"github.com/goharbor/harbor/src/pkg/version"
)

const defaultRootCert = "/etc/core/ca/ca.crt"

// for UT only
var testRootCertPath = ""

// Ctl is the default instance of the package
var Ctl = NewController()

// Data wraps common systeminfo data
type Data struct {
	AuthMode          string
	SelfRegistration  bool
	HarborVersion     string
	AuthProxySettings *models.HTTPAuthProxy
	Protected         *protectedData
}

type protectedData struct {
	CurrentTime                 time.Time
	WithNotary                  bool
	RegistryURL                 string
	ExtURL                      string
	ProjectCreationRestrict     string
	HasCARoot                   bool
	RegistryStorageProviderName string
	ReadOnly                    bool
	WithChartMuseum             bool
	NotificationEnable          bool
}

// Options provide a set of attributes to control what info should be returned
type Options struct {
	// WithProtectedInfo controls if the protected info, which are considered to be sensitive, should be returned
	WithProtectedInfo bool
}

// Controller defines the methods needed for systeminfo API
type Controller interface {

	// GetInfo consolidates the info of the system by checking settings in DB and env vars
	GetInfo(ctx context.Context, opt Options) (*Data, error)

	// GetCapacity returns total and free space of the storage in byte
	GetCapacity(ctx context.Context) (*imagestorage.Capacity, error)

	// GetCA returns a ReadCloser of Harbor's CA if it's configured and accessible from Harbor core
	GetCA(ctx context.Context) (io.ReadCloser, error)
}

type controller struct{}

func (c *controller) GetInfo(ctx context.Context, opt Options) (*Data, error) {
	logger := log.GetLogger(ctx)
	cfg, err := config.GetSystemCfg(ctx)
	if err != nil {
		logger.Errorf("Error occurred getting config: %v", err)
		return nil, err
	}
	res := &Data{
		AuthMode:         utils.SafeCastString(cfg[common.AUTHMode]),
		SelfRegistration: utils.SafeCastBool(cfg[common.SelfRegistration]),
		HarborVersion:    fmt.Sprintf("%s-%s", version.ReleaseVersion, version.GitCommit),
	}
	if res.AuthMode == common.HTTPAuth {
		if s, err := config.HTTPAuthProxySetting(ctx); err == nil {
			res.AuthProxySettings = s
		} else {
			logger.Warningf("Failed to get auth proxy setting, error: %v", err)
		}
	}

	if !opt.WithProtectedInfo {
		return res, nil
	}
	extURL := cfg[common.ExtEndpoint].(string)
	var registryURL string
	if l := strings.Split(extURL, "://"); len(l) > 1 {
		registryURL = l[1]
	} else {
		registryURL = l[0]
	}
	_, caStatErr := os.Stat(defaultRootCert)
	enableCADownload := caStatErr == nil && strings.HasPrefix(extURL, "https://")
	res.Protected = &protectedData{
		CurrentTime:                 time.Now(),
		WithNotary:                  config.WithNotary(),
		WithChartMuseum:             config.WithChartMuseum(),
		ReadOnly:                    config.ReadOnly(ctx),
		ExtURL:                      extURL,
		RegistryURL:                 registryURL,
		HasCARoot:                   enableCADownload,
		ProjectCreationRestrict:     utils.SafeCastString(cfg[common.ProjectCreationRestriction]),
		RegistryStorageProviderName: utils.SafeCastString(cfg[common.RegistryStorageProviderName]),
		NotificationEnable:          utils.SafeCastBool(cfg[common.NotificationEnable]),
	}
	return res, nil
}

func (c *controller) GetCapacity(ctx context.Context) (*imagestorage.Capacity, error) {
	systeminfo.Init()
	return imagestorage.GlobalDriver.Cap()
}

func (c *controller) GetCA(ctx context.Context) (io.ReadCloser, error) {
	logger := log.GetLogger(ctx)
	path := defaultRootCert
	if len(testRootCertPath) > 0 {
		path = testRootCertPath
	}
	if _, err := os.Stat(path); err == nil {
		return os.Open(path)
	} else if os.IsNotExist(err) {
		return nil, errors.NotFoundError(fmt.Errorf("cert not found in path: %s", path))
	} else {
		logger.Errorf("Failed to stat the cert, path: %s, error: %v", path, err)
		return nil, err
	}
}

// NewController return an instance of controller
func NewController() Controller {
	return &controller{}
}
