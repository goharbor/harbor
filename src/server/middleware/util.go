package middleware

import (
	"context"
)

type contextKey string

const (
	// manifestInfoKey the context key for manifest info
	manifestInfoKey = contextKey("ManifestInfo")
)

// ManifestInfo ...
type ManifestInfo struct {
	ProjectID  int64
	Repository string
	Tag        string
	Digest     string
}

// NewManifestInfoContext returns context with manifest info
func NewManifestInfoContext(ctx context.Context, info *ManifestInfo) context.Context {
	return context.WithValue(ctx, manifestInfoKey, info)
}

// ManifestInfoFromContext returns manifest info from context
func ManifestInfoFromContext(ctx context.Context) (*ManifestInfo, bool) {
	info, ok := ctx.Value(manifestInfoKey).(*ManifestInfo)
	return info, ok
}
