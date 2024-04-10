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
