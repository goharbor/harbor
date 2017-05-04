package controllers

import (
	"strings"

	"github.com/astaxie/beego"
	"github.com/vmware/harbor/src/ui/proxy"
)

// RegistryProxy is the endpoint on UI for a reverse proxy pointing to registry
type RegistryProxy struct {
	beego.Controller
}

// Handle is the only entrypoint for incoming requests, all requests must go through this func.
func (p *RegistryProxy) Handle() {
	req := p.Ctx.Request
	rw := p.Ctx.ResponseWriter
	req.URL.Path = strings.TrimPrefix(req.URL.Path, proxy.RegistryProxyPrefix)
	//TODO interceptors
	proxy.Proxy.ServeHTTP(rw, req)
}

// Render ...
func (p *RegistryProxy) Render() error {
	return nil
}
