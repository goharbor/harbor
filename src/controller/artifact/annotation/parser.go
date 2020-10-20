package annotation

import (
	"context"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	reg "github.com/goharbor/harbor/src/pkg/registry"
)

const (
	// GIF is icon content type image/gif
	GIF = "image/gif"
	// PNG is icon content type image/png
	PNG = "image/png"
	// JPEG is icon content type image/jpeg
	JPEG = "image/jpeg"

	// AnnotationPrefix is the prefix of annotation
	AnnotationPrefix = "io.goharbor.artifact"

	// SkipList is the key word of skip-list annotation
	SkipList = "skip-list"
	// Icon is the key word of icon annotation
	Icon = "icon"
)

var (
	// registry for registered annotation parsers
	registry = map[string]Parser{}

	// sortedAnnotationVersionList define the order of AnnotationParser from low to high version.
	// Low version annotation parser will parser annotation first.
	sortedAnnotationVersionList = make([]string, 0)
)

func init() {
	v1alpha1Parser := &v1alpha1Parser{
		regCli: reg.Cli,
	}
	RegisterAnnotationParser(v1alpha1Parser, V1alpha1)
}

// NewParser creates a new annotation parser
func NewParser() Parser {
	return &parser{}
}

// Parser parses annotations in artifact manifest
type Parser interface {
	// Parse parses annotations in artifact manifest, abstracts data from artifact config layer into the artifact model
	Parse(ctx context.Context, artifact *artifact.Artifact, manifest []byte) (err error)
}

type parser struct{}

func (p *parser) Parse(ctx context.Context, artifact *artifact.Artifact, manifest []byte) (err error) {
	for _, annotationVersion := range sortedAnnotationVersionList {
		err = GetAnnotationParser(annotationVersion).Parse(ctx, artifact, manifest)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterAnnotationParser register annotation parser
func RegisterAnnotationParser(parser Parser, version string) {
	registry[version] = parser
	sortedAnnotationVersionList = append(sortedAnnotationVersionList, version)
	log.Infof("the annotation parser to parser artifact annotation version %s registered", version)
}

// GetAnnotationParser register annotation parser
func GetAnnotationParser(version string) Parser {
	return registry[version]
}
