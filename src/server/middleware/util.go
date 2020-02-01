package middleware

import (
	"context"
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
	"regexp"
)

type contextKey string

const (
	// RepositorySubexp is the name for sub regex that maps to repository name in the url
	RepositorySubexp = "repository"
	// ReferenceSubexp is the name for sub regex that maps to reference (tag or digest) url
	ReferenceSubexp = "reference"
	// DigestSubexp is the name for sub regex that maps to digest in the url
	DigestSubexp = "digest"
	// ArtifactInfoKey the context key for artifact info
	ArtifactInfoKey = contextKey("artifactInfo")
	// manifestInfoKey the context key for manifest info
	manifestInfoKey = contextKey("ManifestInfo")
	// ScannerPullCtxKey the context key for robot account to bypass the pull policy check.
	ScannerPullCtxKey = contextKey("ScannerPullCheck")
	// SkipInjectRegistryCredKey is the context key telling registry proxy to skip adding credentials
	SkipInjectRegistryCredKey = contextKey("SkipInjectRegistryCredential")
)

var (
	// V2ManifestURLRe is the regular expression for matching request v2 handler to view/delete manifest
	V2ManifestURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/manifests/(?P<%s>%s|%s)$`, RepositorySubexp, reference.NameRegexp.String(), ReferenceSubexp, reference.TagRegexp.String(), digest.DigestRegexp.String()))
	// V2TagListURLRe is the regular expression for matching request to v2 handler to list tags
	V2TagListURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/tags/list`, RepositorySubexp, reference.NameRegexp.String()))
	// V2BlobURLRe is the regular expression for matching request to v2 handler to retrieve delete a blob
	V2BlobURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/blobs/(?P<%s>%s)$`, RepositorySubexp, reference.NameRegexp.String(), DigestSubexp, digest.DigestRegexp.String()))
	// V2BlobUploadURLRe is the regular expression for matching the request to v2 handler to upload a blob, the upload uuid currently is not put into a group
	V2BlobUploadURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/blobs/uploads[/a-zA-Z0-9\-_\.=]*$`, RepositorySubexp, reference.NameRegexp.String()))
	// V2CatalogURLRe is the regular expression for mathing the request to v2 handler to list catalog
	V2CatalogURLRe = regexp.MustCompile(`^/v2/_catalog$`)
)

// ManifestInfo ...
type ManifestInfo struct {
	ProjectID  int64
	Repository string
	Tag        string
	Digest     string
}

// ArtifactInfo ...
type ArtifactInfo struct {
	Repository           string
	Reference            string
	ProjectName          string
	Digest               string
	BlobMountRepository  string
	BlobMountProjectName string
	BlobMountDigest      string
}

// ArtifactInfoFromContext returns the artifact info from context
func ArtifactInfoFromContext(ctx context.Context) (*ArtifactInfo, bool) {
	info, ok := ctx.Value(ArtifactInfoKey).(*ArtifactInfo)
	return info, ok
}

// SkipInjectRegistryCred reflects whether the inject credentials should be skipped
func SkipInjectRegistryCred(ctx context.Context) bool {
	res, ok := ctx.Value(SkipInjectRegistryCredKey).(bool)
	return ok && res
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

// NewScannerPullContext returns context with policy check info
func NewScannerPullContext(ctx context.Context, scannerPull bool) context.Context {
	return context.WithValue(ctx, ScannerPullCtxKey, scannerPull)
}
