package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/assert"
)

func TestMLogger_Verbose(t *testing.T) {
	l := newMigrateLogger()
	if log.DefaultLogger().GetLevel() <= log.DebugLevel {
		assert.True(t, l.Verbose())
	} else {
		assert.False(t, l.Verbose())
	}
}
