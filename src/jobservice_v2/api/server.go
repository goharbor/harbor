// Copyright 2018 The Harbor Authors. All rights reserved.

package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/config"
	"github.com/vmware/harbor/src/jobservice_v2/env"
)

//Server serves the http requests.
type Server struct {
	//The real backend http server to serve the requests
	httpServer *http.Server

	//Define the routes of http service
	router Router

	//Keep the configurations of server
	config ServerConfig

	//The context
	context *env.Context
}

//ServerConfig contains the configurations of Server.
type ServerConfig struct {
	//Protocol server listening on: https/http
	Protocol string

	//Server listening port
	Port uint

	//Cert file path if using https
	Cert string

	//Key file path if using https
	Key string
}

//NewServer is constructor of Server.
func NewServer(ctx *env.Context, router Router, cfg ServerConfig) *Server {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      http.HandlerFunc(router.ServeHTTP),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	//Initialize TLS/SSL config if protocol is https
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

	return &Server{
		httpServer: srv,
		router:     router,
		config:     cfg,
		context:    ctx,
	}
}

//Start the server to serve requests.
func (s *Server) Start() {
	s.context.WG.Add(1)

	go func() {
		var err error

		defer func() {
			if s.context.WG != nil {
				s.context.WG.Done()
			}
		}()
		if s.config.Protocol == config.JobServiceProtocolHTTPS {
			err = s.httpServer.ListenAndServeTLS(s.config.Cert, s.config.Key)
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil {
			s.context.ErrorChan <- err
		}
	}()
}

//Stop server gracefully.
func (s *Server) Stop() {
	go func() {
		shutDownCtx, cancel := context.WithTimeout(s.context.SystemContext, 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutDownCtx); err != nil {
			log.Errorf("Shutdown API server failed with error: %s\n", err)
		} else {
			log.Infof("API server is gracefully shutdown")
		}
	}()
}
