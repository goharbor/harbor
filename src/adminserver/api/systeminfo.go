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

package api

import (
	"net/http"

	"github.com/vmware/harbor/src/adminserver/systeminfo/imagestorage"
	"github.com/vmware/harbor/src/common/utils/log"
)

// Capacity handles /api/systeminfo/capacity and returns system capacity
func Capacity(w http.ResponseWriter, r *http.Request) {
	capacity, err := imagestorage.GlobalDriver.Cap()
	if err != nil {
		log.Errorf("failed to get capacity: %v", err)
		handleInternalServerError(w)
		return
	}

	if err = writeJSON(w, capacity); err != nil {
		log.Errorf("failed to write response: %v", err)
		return
	}
}
