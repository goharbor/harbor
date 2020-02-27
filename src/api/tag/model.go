package tag

import (
	"github.com/goharbor/harbor/src/pkg/signature"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
)

// Tag is the overall view of tag
type Tag struct {
	tag.Tag
	Immutable bool `json:"immutable"`
	Signed    bool `json:"signed"`
}

// Option is used to specify the properties returned when listing/getting tags
type Option struct {
	WithImmutableStatus bool
	WithSignature       bool
	SignatureChecker    *signature.Checker
}
