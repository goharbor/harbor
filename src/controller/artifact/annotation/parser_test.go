package annotation

import (
	"testing"

	fp "github.com/goharbor/harbor/src/testing/pkg/parser"

	"github.com/stretchr/testify/suite"
)

type parserTestSuite struct {
	suite.Suite
}

func (p *parserTestSuite) SetupTest() {
	registry = map[string]Parser{}
}

func (p *parserTestSuite) TestRegisterAnnotationParser() {
	// success
	version := "v1alpha1"
	parser := &fp.Parser{}
	RegisterAnnotationParser(parser, version)
	p.Equal(map[string]Parser{version: parser}, registry)
}

func (p *parserTestSuite) TestGetAnnotationParser() {
	// register the parser
	version := "v1alpha1"
	RegisterAnnotationParser(&fp.Parser{}, "v1alpha1")

	// get the parser
	parser := GetAnnotationParser(version)
	p.Require().NotNil(parser)
	_, ok := parser.(*fp.Parser)
	p.True(ok)
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, &parserTestSuite{})
}
