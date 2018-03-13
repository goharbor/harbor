// Copyright 2018 The Harbor Authors. All rights reserved.

package job

import (
	"context"

	hlog "github.com/vmware/harbor/src/common/utils/log"
)

//Context is combination of BaseContext and other job specified resources.
//Context will be the real execution context for one job.
//Use pointer to point to the singleton BaseContext copy.
type Context struct {
	//System context
	SystemContext context.Context

	//Logger for job
	Logger *hlog.Logger
}
