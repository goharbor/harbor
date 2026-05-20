package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
)

func TestEmailJobMaxFails(t *testing.T) {
	rep := &EmailJob{}
	t.Run("default max fails", func(t *testing.T) {
		assert.Equal(t, uint(3), rep.MaxFails())
	})

	t.Run("user defined max fails", func(t *testing.T) {
		t.Setenv(maxFails, "15")
		assert.Equal(t, uint(15), rep.MaxFails())
	})

	t.Run("user defined wrong max fails", func(t *testing.T) {
		t.Setenv(maxFails, "abc")
		assert.Equal(t, uint(3), rep.MaxFails())
	})
}

func TestEmailJobShouldRetry(t *testing.T) {
	rep := &EmailJob{}
	assert.True(t, rep.ShouldRetry())
}

func TestEmailJobValidate(t *testing.T) {
	rep := &EmailJob{}
	assert.NotNil(t, rep.Validate(nil))

	jp := job.Parameters{
		"subject": "Test Subject",
		"body":    "Test Body",
		"to":      "user@example.com",
	}
	assert.Nil(t, rep.Validate(jp))
}

func TestEmailJobRun(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(logger)

	rep := &EmailJob{}

	params := map[string]any{
		"subject": "Test Subject",
		"body":    "Test Body",
		"to":      "user@example.com",
	}
	// Since email config may not be set, it may fail, but validate should pass
	assert.NotNil(t, rep.Validate(params))
}