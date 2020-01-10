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
	"crypto/tls"
	"flag"
	"github.com/docker/distribution/registry/storage/driver/factory"
	regConf "github.com/goharbor/harbor/src/registryctl/config/registry"
	"net/http"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/registryctl/config"
	"github.com/goharbor/harbor/src/registryctl/handlers"

	_ "github.com/docker/distribution/registry/storage/driver/azure"
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
	_ "github.com/docker/distribution/registry/storage/driver/gcs"
	_ "github.com/docker/distribution/registry/storage/driver/inmemory"
	_ "github.com/docker/distribution/registry/storage/driver/middleware/cloudfront"
	_ "github.com/docker/distribution/registry/storage/driver/middleware/redirect"
	_ "github.com/docker/distribution/registry/storage/driver/oss"
	_ "github.com/docker/distribution/registry/storage/driver/s3-aws"
	_ "github.com/docker/distribution/registry/storage/driver/swift"
)

const RegConf = "/etc/registry/config.yml"

// RegistryCtl for registry controller
type RegistryCtl struct {
	ServerConf config.Configuration
	Handler    http.Handler
}

// Start the registry controller
func (s *RegistryCtl) Start() {
	regCtl := &http.Server{
		Addr:    ":" + s.ServerConf.Port,
		Handler: s.Handler,
	}

	if s.ServerConf.Protocol == "HTTPS" {
		tlsCfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		regCtl.TLSConfig = tlsCfg
		regCtl.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}

	var err error
	if s.ServerConf.Protocol == "HTTPS" {
		err = regCtl.ListenAndServeTLS(s.ServerConf.HTTPSConfig.Cert, s.ServerConf.HTTPSConfig.Key)
	} else {
		err = regCtl.ListenAndServe()
	}

	if err != nil {
		log.Fatal(err)
	}

	return
}

func main() {

	configPath := flag.String("c", "", "Specify the yaml rConf file path")
	flag.Parse()

	if configPath == nil || len(*configPath) == 0 {
		flag.Usage()
		log.Fatal("Config file should be specified")
	}

	if err := config.DefaultConfig.Load(*configPath, true); err != nil {
		log.Fatalf("Failed to load configurations with error: %s\n", err)
	}

	regCtl := &RegistryCtl{
		ServerConf: *config.DefaultConfig,
		Handler:    handlers.NewHandlerChain(),
	}

	// set the global driver
	rConf, err := regConf.ResolveConfiguration(RegConf)
	if err != nil {
		log.Error(err)
		return
	}
	regConf.StorageDriver, err = factory.Create(rConf.Storage.Type(), rConf.Storage.Parameters())
	if err != nil {
		log.Error(err)
		return
	}

	regCtl.Start()
}
