package notifications

import "github.com/goharbor/harbor/src/core/api"

// BaseHandler extracts the common funcs, all notification handlers should shadow this struct
type BaseHandler struct {
	api.BaseController
}

// Prepare disable the xsrf as the request is from other components and do not require the xsrf token
func (bh *BaseHandler) Prepare() {
	bh.EnableXSRF = false
}
