package content

import "github.com/containerd/containerd/content"

// ProvideIngester is the interface that groups the basic Read and Write methods.
type ProvideIngester interface {
	content.Provider
	content.Ingester
}
