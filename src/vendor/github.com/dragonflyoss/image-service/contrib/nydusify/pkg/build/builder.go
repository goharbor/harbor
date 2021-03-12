// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type BuilderOption struct {
	ParentBootstrapPath string
	BootstrapPath       string
	RootfsPath          string
	BackendType         string
	BackendConfig       string
	PrefetchDir         string
	WhiteoutSpec        string
	OutputJSONPath      string
}

type Builder struct {
	binaryPath string
	stdout     io.Writer
	stderr     io.Writer
}

func NewBuilder(binaryPath string) *Builder {
	return &Builder{
		binaryPath: binaryPath,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
	}
}

// Run exec nydus-image CLI to build layer
func (builder *Builder) Run(option BuilderOption) error {
	var args []string
	if option.ParentBootstrapPath == "" {
		args = []string{
			"create",
		}
	} else {
		args = []string{
			"create",
			"--parent-bootstrap",
			option.ParentBootstrapPath,
		}
	}
	args = append(
		args,
		"--bootstrap",
		option.BootstrapPath,
		"--backend-type",
		option.BackendType,
		"--backend-config",
		option.BackendConfig,
		option.RootfsPath,
		"--log-level",
		"warn",
		"--whiteout-spec",
		option.WhiteoutSpec,
		"--output-json",
		option.OutputJSONPath,
	)

	if option.PrefetchDir != "" {
		args = append(args, "--prefetch-policy", "fs")
	}

	logrus.Debugf("\tCommand: %s %s", builder.binaryPath, strings.Join(args[:], " "))

	cmd := exec.Command(builder.binaryPath, args...)
	cmd.Stdout = builder.stdout
	cmd.Stderr = builder.stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	io.WriteString(stdin, option.PrefetchDir)
	stdin.Close()

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
