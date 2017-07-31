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
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/clair"
	"github.com/vmware/harbor/src/jobservice/config"
	"github.com/vmware/harbor/src/jobservice/utils"

	"fmt"
)

// Initializer will handle the initialise state pull the manifest, prepare token.
type Initializer struct {
	Context *JobContext
}

// Enter ...
func (iz *Initializer) Enter() (string, error) {
	logger := iz.Context.Logger
	logger.Infof("Entered scan initializer")
	regURL, err := config.LocalRegURL()
	if err != nil {
		logger.Errorf("Failed to read regURL, error: %v", err)
		return "", err
	}
	repoClient, err := utils.NewRepositoryClientForJobservice(iz.Context.Repository)
	if err != nil {
		logger.Errorf("An error occurred while creating repository client: %v", err)
		return "", err
	}

	_, _, payload, err := repoClient.PullManifest(iz.Context.Digest, []string{schema2.MediaTypeManifest})
	if err != nil {
		logger.Errorf("Error pulling manifest for image %s:%s :%v", iz.Context.Repository, iz.Context.Tag, err)
		return "", err
	}
	manifest, _, err := distribution.UnmarshalManifest(schema2.MediaTypeManifest, payload)
	if err != nil {
		logger.Error("Failed to unMarshal manifest from response")
		return "", err
	}

	tk, err := utils.GetTokenForRepo(iz.Context.Repository)
	if err != nil {
		return "", err
	}
	iz.Context.token = tk
	iz.Context.clairClient = clair.NewClient(config.ClairEndpoint(), logger)
	iz.prepareLayers(regURL, manifest.References())
	return StateScanLayer, nil
}

func (iz *Initializer) prepareLayers(registryEndpoint string, descriptors []distribution.Descriptor) {
	//	logger := iz.Context.Logger
	tokenHeader := map[string]string{"Authorization": fmt.Sprintf("Bearer %s", iz.Context.token)}
	for _, d := range descriptors {
		if d.MediaType == schema2.MediaTypeConfig {
			continue
		}
		l := models.ClairLayer{
			Name:    fmt.Sprintf("%d-%s", iz.Context.JobID, d.Digest),
			Headers: tokenHeader,
			Format:  "Docker",
			Path:    utils.BuildBlobURL(registryEndpoint, iz.Context.Repository, string(d.Digest)),
		}
		if len(iz.Context.layers) > 0 {
			l.ParentName = iz.Context.layers[len(iz.Context.layers)-1].Name
		}
		iz.Context.layers = append(iz.Context.layers, l)
	}
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
	currentLayer := ls.Context.layers[ls.Context.current]
	logger.Infof("Entered scan layer handler, current: %d, layer name: %s", ls.Context.current, currentLayer.Name)
	err := ls.Context.clairClient.ScanLayer(currentLayer)
	if err != nil {
		logger.Errorf("Unexpected error: %v", err)
		return "", err
	}
	ls.Context.current++
	if ls.Context.current == len(ls.Context.layers) {
		return StateSummarize, nil
	}
	logger.Infof("After scanning, return with next state: %s", StateScanLayer)
	return StateScanLayer, nil
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
	layerName := sh.Context.layers[len(sh.Context.layers)-1].Name
	logger.Infof("Top layer's name: %s, will use it to get the vulnerability result of image", layerName)
	if err := clair.UpdateScanOverview(sh.Context.Digest, layerName); err != nil {
		return "", nil
	}
	return models.JobFinished, nil
}

// Exit ...
func (sh *SummarizeHandler) Exit() error {
	return nil
}
