package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSupportedMimeType(t *testing.T) {
	// Test with a supported mime type
	assert.True(t, isSupportedMimeType(MimeTypeSBOMReport), "isSupportedMimeType should return true for supported mime types")

	// Test with an unsupported mime type
	assert.False(t, isSupportedMimeType("unsupported/mime-type"), "isSupportedMimeType should return false for unsupported mime types")
}

func TestConvertCapability(t *testing.T) {
	md := &ScannerAdapterMetadata{
		Capabilities: []*ScannerCapability{
			{Type: ScanTypeSbom},
			{Type: ScanTypeVulnerability},
		},
	}
	result := md.ConvertCapability()
	assert.Equal(t, result[supportSBOM], true)
	assert.Equal(t, result[supportVulnerability], true)
}

func TestConvertCapabilityOldScaner(t *testing.T) {
	md := &ScannerAdapterMetadata{
		Capabilities: []*ScannerCapability{
			{
				ConsumesMimeTypes: []string{"application/vnd.oci.image.manifest.v1+json", "application/vnd.docker.distribution.manifest.v2+json"},
				ProducesMimeTypes: []string{MimeTypeNativeReport},
			},
		},
	}
	result := md.ConvertCapability()
	assert.Equal(t, result[supportSBOM], false)
	assert.Equal(t, result[supportVulnerability], true)
}
