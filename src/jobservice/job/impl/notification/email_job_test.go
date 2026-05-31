package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/jobservice/job"
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
		"address": "smtp.example.com",
		"from":    "harbor@example.com",
	}
	assert.Nil(t, rep.Validate(jp))
}

func TestEmailJobRun(t *testing.T) {
	rep := &EmailJob{}

	params := map[string]any{
		"subject": "Test Subject",
		"body":    "Test Body",
		"to":      "user@example.com",
		"address": "smtp.example.com",
		"from":    "harbor@example.com",
	}
	assert.Nil(t, rep.Validate(params))
}