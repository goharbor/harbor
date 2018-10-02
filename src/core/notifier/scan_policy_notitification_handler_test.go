package notifier

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanPolicyNotificationHandler(t *testing.T) {
	assert := assert.New(t)
	s := &ScanPolicyNotificationHandler{}
	assert.True(s.IsStateful())
	err := s.Handle("")
	if assert.NotNil(err) {
		assert.Contains(err.Error(), "invalid type")
	}
}
