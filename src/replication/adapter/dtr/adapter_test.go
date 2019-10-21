package dtr

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	a := &adapter{}
	info, err := a.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.EqualValues(t, 1, len(info.SupportedResourceTypes))
	assert.EqualValues(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}
