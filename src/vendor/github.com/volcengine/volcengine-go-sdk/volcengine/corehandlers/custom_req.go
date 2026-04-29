package corehandlers

import (
	"github.com/volcengine/volcengine-go-sdk/volcengine/custom"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

var CustomerRequestHandler = request.NamedHandler{
	Name: "core.CustomerRequestHandler",
	Fn: func(r *request.Request) {
		if r.Config.ExtendHttpRequest != nil {
			r.Config.ExtendHttpRequest(r.Context(), r.HTTPRequest)
		}

		if r.Config.ExtendHttpRequestWithMeta != nil {
			r.Config.ExtendHttpRequestWithMeta(r.Context(), r.HTTPRequest, custom.RequestMetadata{
				ServiceName: r.ClientInfo.ServiceName,
				Version:     r.ClientInfo.APIVersion,
				Action:      r.Operation.Name,
				HttpMethod:  r.Operation.HTTPMethod,
				Region:      *r.Config.Region,
			})
		}
	},
}
