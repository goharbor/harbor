package proxy

import (
	"github.com/vmware/harbor/src/ui/config"

	"fmt"
	"net/http/httputil"
	"net/url"
)

// Proxy is the instance of the reverse proxy in this package.
var Proxy *httputil.ReverseProxy

// RegistryProxyPrefix is the prefix of url on UI.
const RegistryProxyPrefix = "/registryproxy"

// Init initialize the Proxy instance.
func Init(urls ...string) error {
	var err error
	var registryURL string
	if len(urls) > 1 {
		return fmt.Errorf("the parm, urls should have only 0 or 1 elements")
	}
	if len(urls) == 0 {
		registryURL, err = config.RegistryURL()
		if err != nil {
			return err
		}
	} else {
		registryURL = urls[0]
	}
	targetURL, err := url.Parse(registryURL)
	if err != nil {
		return err
	}
	Proxy = httputil.NewSingleHostReverseProxy(targetURL)
	return nil
}

//func StartProxy(registryURL string) {
//http.ListenAndServe(":5000", Proxy)
//}
