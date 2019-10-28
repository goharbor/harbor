package controllers

import (
	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/core/middlewares"
)

// RegistryProxy is the endpoint on UI for a reverse proxy pointing to registry
type RegistryProxy struct {
	beego.Controller
}

// Prepare turn off the xsrf check for registry proxy
func (p *RegistryProxy) Prepare() {
	p.EnableXSRF = false
}

// Handle is the only entrypoint for incoming requests, all requests must go through this func.
func (p *RegistryProxy) Handle() {
	req := p.Ctx.Request
	rw := p.Ctx.ResponseWriter
	middlewares.Handle(rw, req)
}

// Render ...
func (p *RegistryProxy) Render() error {
	return nil
}
