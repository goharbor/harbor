package job

import (
	hlog "github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/core"
)

//Context is combination of BaseContext and other job specified resources.
//Context will be the real execution context for one job.
//Use pointer to point to the singleton BaseContext copy.
type Context struct {
	//Base context
	*core.BaseContext

	//Logger for job
	Logger *hlog.Logger
}
