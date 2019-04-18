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

package api

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"context"
	"github.com/goharbor/harbor/src/jobservice/config"
)

// Server serves the http requests.
type Server struct {
	// The real backend http server to serve the requests
	httpServer *http.Server

	// Define the routes of http service
	router Router

	// Keep the configurations of server
	config ServerConfig

	// The context
	context context.Context
}

// ServerConfig contains the configurations of Server.
type ServerConfig struct {
	// Protocol server listening on: https/http
	Protocol string

	// Server listening port
	Port uint

	// Cert file path if using https
	Cert string

	// Key file path if using https
	Key string
}

// NewServer is constructor of Server.
func NewServer(ctx context.Context, router Router, cfg ServerConfig) *Server {
	apiServer := &Server{
		router:  router,
		config:  cfg,
		context: ctx,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      http.HandlerFunc(router.ServeHTTP),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Initialize TLS/SSL config if protocol is https
	if cfg.Protocol == config.JobServiceProtocolHTTPS {
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

		srv.TLSConfig = tlsCfg
		srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}

	apiServer.httpServer = srv

	return apiServer
}

// Start the server to serve requests.
// Blocking call
func (s *Server) Start() error {
	if s.config.Protocol == config.JobServiceProtocolHTTPS {
		return s.httpServer.ListenAndServeTLS(s.config.Cert, s.config.Key)
	} else {
		return s.httpServer.ListenAndServe()
	}
}

// Stop server gracefully.
func (s *Server) Stop() error {
	shutDownCtx, cancel := context.WithTimeout(s.context, 15*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(shutDownCtx)
}
