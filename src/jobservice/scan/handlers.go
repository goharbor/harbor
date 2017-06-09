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

package scan

import (
	"github.com/vmware/harbor/src/common/models"
)

// Initializer will handle the initialise state pull the manifest, prepare token.
type Initializer struct {
	Context *JobContext
}

// Enter ...
func (iz *Initializer) Enter() (string, error) {
	logger := iz.Context.Logger
	logger.Infof("Entered scan initializer")
	return StateScanLayer, nil
}

// Exit ...
func (iz *Initializer) Exit() error {
	return nil
}

//LayerScanHandler will call clair API to trigger scanning.
type LayerScanHandler struct {
	Context *JobContext
}

// Enter ...
func (ls *LayerScanHandler) Enter() (string, error) {
	logger := ls.Context.Logger
	logger.Infof("Entered scan layer handler")
	return StateSummarize, nil
}

// Exit ...
func (ls *LayerScanHandler) Exit() error {
	return nil
}

// SummarizeHandler will summarize the vulnerability and feature information of Clair, and store into Harbor's DB.
type SummarizeHandler struct {
	Context *JobContext
}

// Enter ...
func (sh *SummarizeHandler) Enter() (string, error) {
	logger := sh.Context.Logger
	logger.Infof("Entered summarize handler")
	return models.JobFinished, nil
}

// Exit ...
func (sh *SummarizeHandler) Exit() error {
	return nil
}
