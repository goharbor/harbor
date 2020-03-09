package notifications

import "github.com/goharbor/harbor/src/core/api"

// BaseHandler extracts the common funcs, all notification handlers should shadow this struct
type BaseHandler struct {
	api.BaseController
}
