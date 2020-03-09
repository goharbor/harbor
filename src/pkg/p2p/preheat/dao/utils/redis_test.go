package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisAddr(t *testing.T) {
	input := "127.0.0.1:6379,100,,0,30"
	expect := "redis://127.0.0.1:6379/0"
	output, ok := RedisAddr(input)
	assert.Equal(t, expect, output)
	assert.True(t, ok)

	input = "127.0.0.1:6379,100,pwd,0,30"
	expect = "redis://arbitrary_username:pwd@127.0.0.1:6379/0"
	output, ok = RedisAddr(input)
	assert.Equal(t, expect, output)
	assert.True(t, ok)
}
