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
	"net/http"
	"os"

	"github.com/goharbor/harbor/src/adminserver/handlers"
	syscfg "github.com/goharbor/harbor/src/adminserver/systemcfg"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// Server for admin component
type Server struct {
	Port    string
	TLSCert string
	TLSKey  string
	Handler http.Handler
}

// Serve the API
func (s *Server) Serve() error {
	server := &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.Handler,
	}

	if len(s.TLSCert) >= 0 && len(s.TLSKey) > 0 {
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

		server.TLSConfig = tlsCfg
		server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)

		return server.ListenAndServeTLS(s.TLSCert, s.TLSKey)
	}
	return server.ListenAndServe()

}

func main() {
	log.Info("initializing system configurations...")
	if err := syscfg.Init(); err != nil {
		log.Fatalf("failed to initialize the system: %v", err)
	}
	log.Info("system initialization completed")

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "80"
	}

	tlscert := os.Getenv("TLS_CERT")
	tlskey := os.Getenv("TLS_KEY")

	server := &Server{
		Port:    port,
		TLSCert: tlscert,
		TLSKey:  tlskey,
		Handler: handlers.NewHandler(),
	}
	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
