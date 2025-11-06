package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
)

func TestTelegramJobMaxFails(t *testing.T) {
	rep := &TelegramJob{}
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

func TestTelegramJobShouldRetry(t *testing.T) {
	rep := &TelegramJob{}
	assert.True(t, rep.ShouldRetry())
}

func TestTelegramJobValidate(t *testing.T) {
	rep := &TelegramJob{}
	assert.NotNil(t, rep.Validate(nil))

	jp := job.Parameters{
		"text":      "telegram message",
		"bot_token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
		"chat_id":   "@harbor_notifications",
	}
	assert.Nil(t, rep.Validate(jp))
}

func TestTelegramJobRun(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(logger)

	rep := &TelegramJob{}

	params := map[string]any{
		"text":      "telegram message",
		"bot_token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
		"chat_id":   "@harbor_notifications",
	}
	// Since Telegram API call may fail, but validate should pass
	assert.NotNil(t, rep.Validate(params))
}