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

package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vmware/harbor/src/adminserver/api"
)

func newRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/api/configurations", api.ListCfgs).Methods("GET")
	r.HandleFunc("/api/configurations", api.UpdateCfgs).Methods("PUT")
	r.HandleFunc("/api/configurations/reset", api.ResetCfgs).Methods("POST")
	r.HandleFunc("/api/systeminfo/capacity", api.Capacity).Methods("GET")
	return r
}
