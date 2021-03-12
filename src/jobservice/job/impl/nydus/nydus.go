package nydus

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter/provider"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"strings"
)

type NydusifyConverter struct {
	repository string
	tag        string
	username   string
	password   string
	coreUrl    string
	logger     logger.Interface
}

// MaxFails implements the interface in job/Interface
func (n *NydusifyConverter) MaxFails() uint {
	return 1
}

// MaxCurrency is implementation of same method in Interface.
func (n *NydusifyConverter) MaxCurrency() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (n *NydusifyConverter) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (n *NydusifyConverter) Validate(params job.Parameters) error {
	return nil
}

// init ...
func (n *NydusifyConverter) init(ctx job.Context, params job.Parameters) error {
	n.logger = ctx.GetLogger()
	n.coreUrl = strings.TrimPrefix(params["core_url"].(string), "http://")
	n.username = params["username"].(string)
	n.password = params["password"].(string)
	n.repository = params["repository"].(string)
	n.tag = params["tag"].(string)
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Run implements the interface in job/Interface
func (n *NydusifyConverter) Run(ctx job.Context, params job.Parameters) error {
	n.init(ctx, params)
	jLog := ctx.GetLogger()

	// TODO needs to define these two parameters.
	wordDir := "/var/log/jobs/nydus-tmp"
	nydusImagePath := "/harbor/nydus-image"

	source := fmt.Sprintf("%s/%s:%s", n.coreUrl, n.repository, n.tag)
	target := fmt.Sprintf("%s/%s:%s-nydus", n.coreUrl, n.repository, n.tag)
	auth := basicAuth(n.username, n.password)
	insecure := true
	jLog.Info(target)

	logger, err := provider.DefaultLogger()
	if err != nil {
		return err
	}

	// Create remote with auth string for registry communication
	sourceRemote, err := provider.DefaultRemoteWithAuth(source, insecure, auth)
	if err != nil {
		jLog.Info(err)
		return err
	}

	targetRemote, err := provider.DefaultRemoteWithAuth(target, insecure, auth)
	if err != nil {
		jLog.Info(err)
		return err
	}

	// Source provider gets source image manifest, config, and layer
	sourceProvider, err := provider.DefaultSource(context.Background(), sourceRemote, wordDir)
	if err != nil {
		jLog.Info(err)
		return err
	}

	opt := converter.Opt{
		Logger:         logger,
		SourceProvider: sourceProvider,
		TargetRemote:   targetRemote,

		WorkDir:        wordDir,
		PrefetchDir:    "/",
		NydusImagePath: nydusImagePath,
		MultiPlatform:  false,
		DockerV2Format: true,
		WhiteoutSpec:   "oci",
	}

	cvt, err := converter.New(opt)
	if err != nil {
		jLog.Info(err)
		return err
	}

	if err := cvt.Convert(context.Background()); err != nil {
		jLog.Info(err)
		return err
	}

	return nil
}
