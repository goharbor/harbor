package request

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/custom"
)

// setContext updates the Request to use the passed in context for cancellation.
// Context will also be used for request retry delay.
//
// Creates shallow copy of the http.Request with the WithContext method.
func setRequestContext(r *Request, ctx volcengine.Context) {
	if r.Config.ExtendContextWithMeta != nil {
		newCtx := r.Config.ExtendContextWithMeta(ctx, custom.RequestMetadata{
			ServiceName: r.ClientInfo.ServiceName,
			Version:     r.ClientInfo.APIVersion,
			Action:      r.Operation.Name,
			HttpMethod:  r.Operation.HTTPMethod,
			Region:      *r.Config.Region,
		})
		r.context = newCtx
		r.HTTPRequest = r.HTTPRequest.WithContext(newCtx)
	} else {
		r.context = ctx
		r.HTTPRequest = r.HTTPRequest.WithContext(ctx)
	}

}
