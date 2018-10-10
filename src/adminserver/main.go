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
	"net/http"
	"os"

	"github.com/goharbor/harbor/src/adminserver/handlers"
	syscfg "github.com/goharbor/harbor/src/adminserver/systemcfg"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// Server for admin component
type Server struct {
	Port    string
	Handler http.Handler
}

// Serve the API
func (s *Server) Serve() error {
	server := &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.Handler,
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
	server := &Server{
		Port:    port,
		Handler: handlers.NewHandler(),
	}
	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
