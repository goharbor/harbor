package proxy

import "github.com/prometheus/client_golang/prometheus"

var (
	ProxyRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_core_registry_requests_total",
			Help: "Total number of registry requests received by the proxy.",
		},

		[]string{"project", "repo", "method"},
	)

	ProxyUpstreamRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_core_proxy_upstream_requests_total",
			Help: "Total number of proxy requests that were sent to the upstream server.",
		},
		[]string{"project", "repo", "method"},
	)
)

func init() {
	prometheus.MustRegister(ProxyRequestsTotal)
	prometheus.MustRegister(ProxyUpstreamRequestsTotal)
}
