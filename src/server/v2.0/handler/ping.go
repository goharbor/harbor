package handler

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/ping"
)

type pingAPI struct {
	BaseAPI
}

func newPingAPI() *pingAPI {
	return &pingAPI{}
}

func (p *pingAPI) GetPing(ctx context.Context, params ping.GetPingParams) middleware.Responder {
	return ping.NewGetPingOK().WithPayload("Pong")
}
