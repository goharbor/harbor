// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type WorkflowOption struct {
	TargetDir      string
	NydusImagePath string
	PrefetchDir    string
	WhiteoutSpec   string
}

type Workflow struct {
	WorkflowOption
	bootstrapPath       string
	blobsDir            string
	backendConfig       string
	parentBootstrapPath string
	builder             *Builder
	debugJSONPath       string
	lastBlobID          string
}

type debugJSON struct {
	Blobs []string
}

// Get latest built blob from blobs directory
func (workflow *Workflow) getLatestBlobPath() (string, error) {
	var data debugJSON
	jsonBytes, err := ioutil.ReadFile(workflow.debugJSONPath)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return "", err
	}
	blobIDs := data.Blobs

	if blobIDs == nil || len(blobIDs) == 0 {
		return "", nil
	}

	latestBlobID := blobIDs[len(blobIDs)-1]
	if latestBlobID != workflow.lastBlobID {
		workflow.lastBlobID = latestBlobID
		blobPath := filepath.Join(workflow.blobsDir, latestBlobID)
		if _, err := os.Stat(blobPath); err == nil {
			return blobPath, nil
		}
	}

	return "", nil
}

// NewWorkflow prepare bootstrap and blobs path for layered build workflow
func NewWorkflow(option WorkflowOption) (*Workflow, error) {
	blobsDir := filepath.Join(option.TargetDir, "blobs")
	if err := os.RemoveAll(blobsDir); err != nil {
		return nil, errors.Wrap(err, "Remove blob directory")
	}
	if err := os.MkdirAll(blobsDir, 0755); err != nil {
		return nil, errors.Wrap(err, "Create blob directory")
	}

	backendConfig := fmt.Sprintf(`{"dir": "%s"}`, blobsDir)
	builder := NewBuilder(option.NydusImagePath)

	debugJSONPath := filepath.Join(option.TargetDir, "output.json")

	return &Workflow{
		WorkflowOption: option,
		blobsDir:       blobsDir,
		backendConfig:  backendConfig,
		builder:        builder,
		debugJSONPath:  debugJSONPath,
	}, nil
}

// Build nydus bootstrap and blob, returned blobPath's basename is sha256 hex string
func (workflow *Workflow) Build(
	layerDir, parentBootstrapPath, bootstrapPath string,
) (string, error) {
	workflow.bootstrapPath = bootstrapPath

	if parentBootstrapPath != "" {
		workflow.parentBootstrapPath = parentBootstrapPath
	}

	if err := workflow.builder.Run(BuilderOption{
		ParentBootstrapPath: workflow.parentBootstrapPath,
		BootstrapPath:       workflow.bootstrapPath,
		RootfsPath:          layerDir,
		BackendType:         "localfs",
		BackendConfig:       workflow.backendConfig,
		PrefetchDir:         workflow.PrefetchDir,
		WhiteoutSpec:        workflow.WhiteoutSpec,
		OutputJSONPath:      workflow.debugJSONPath,
	}); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("build layer %s", layerDir))
	}

	workflow.parentBootstrapPath = workflow.bootstrapPath

	blobPath, err := workflow.getLatestBlobPath()
	if err != nil {
		return "", errors.Wrap(err, "get latest blob")
	}

	return blobPath, nil
}
