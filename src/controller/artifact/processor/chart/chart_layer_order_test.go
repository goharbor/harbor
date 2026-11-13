package chart

import (
	"io"
	"strings"
	"testing"

	"github.com/docker/distribution"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	chartserver "github.com/goharbor/harbor/src/pkg/chart"
	chartmock "github.com/goharbor/harbor/src/testing/pkg/chart"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
)

// Test that the chart archive is selected by media type, regardless of layer order.
func TestAbstractAdditionHelmLayerOrder(t *testing.T) {
	const chartDigest = "sha256:9ec67938b2981d8b333af61b3fcf53e26b4f6f5deb7ccda3de4ea116bb373a17"
	const provenanceDigest = "sha256:92c767331c38926678a4951a9444da3f04c8e4845aa74504072846ba1984e89c"
	const chartLayer = `{"mediaType":"application/vnd.cncf.helm.chart.content.v1.tar+gzip","digest":"` + chartDigest + `","size":153406}`
	const provenanceLayer = `{"mediaType":"application/vnd.cncf.helm.chart.provenance.v1.prov","digest":"` + provenanceDigest + `","size":1840}`
	const manifestPrefix = `{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:5d9a7cb1db4cda0a50b6511df81376cd428a7a179206a9bf579af485dbcc5e57","size":907},"layers":[`

	for _, tc := range []struct {
		name   string
		layers string
	}{
		{"provenance first", provenanceLayer + "," + chartLayer},
		{"chart first", chartLayer + "," + provenanceLayer},
		{"chart only", chartLayer},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest, _, err := distribution.UnmarshalManifest(
				v1.MediaTypeImageManifest,
				[]byte(manifestPrefix+tc.layers+`]}`),
			)
			if err != nil {
				t.Fatal(err)
			}

			regCli := &registry.Client{}
			chartOptr := &chartmock.Operator{}
			p := &processor{
				ManifestProcessor: &base.ManifestProcessor{RegCli: regCli},
				chartOperator:     chartOptr,
			}

			regCli.On("PullManifest", mock.Anything, mock.Anything).
				Return(manifest, "", nil).Once()
			regCli.On("PullBlob", mock.Anything, chartDigest).
				Return(int64(5), io.NopCloser(strings.NewReader("chart")), nil).Once()
			chartOptr.On("GetDetails", mock.Anything).
				Return(&chartserver.VersionDetails{
					Files: map[string]string{AdditionTypeValues: "replicas: 1\n"},
				}, nil).Once()

			addition, err := p.AbstractAddition(nil, &artifact.Artifact{}, AdditionTypeValues)
			if err != nil {
				t.Fatal(err)
			}
			if addition == nil || string(addition.Content) != "replicas: 1\n" {
				t.Fatalf("unexpected chart values: %#v", addition)
			}
			regCli.AssertNotCalled(t, "PullBlob", mock.Anything, provenanceDigest)
			regCli.AssertExpectations(t)
			chartOptr.AssertExpectations(t)
		})
	}
}
