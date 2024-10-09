package gc

import (
	"context"
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreNotFound(t *testing.T) {
	var f = func() error {
		return nil
	}
	assert.Nil(t, ignoreNotFound(f))
	f = func() error {
		return errors.New(nil).WithMessage("my error")
	}
	assert.NotNil(t, ignoreNotFound(f))
	f = func() error {
		return errors.New(nil).WithMessage("my error").WithCode(errors.BadRequestCode)
	}
	assert.NotNil(t, ignoreNotFound(f))
	f = func() error {
		return errors.New(nil).WithMessage("my error").WithCode(errors.NotFoundCode)
	}
	assert.Nil(t, ignoreNotFound(f))
}

func TestDivide(t *testing.T) {
	var result int
	var err error
	result, err = divide(1, 10)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)

	result, err = divide(5, 10)
	assert.Nil(t, err)
	assert.Equal(t, 5, result)

	result, err = divide(30, 10)
	assert.Nil(t, err)
	assert.Equal(t, 3, result)

	result, err = divide(33, 10)
	assert.Nil(t, err)
	assert.Equal(t, 3, result)

	result, err = divide(33, 0)
	assert.NotNil(t, err)
}

func TestDelKeys(t *testing.T) {
	// get redis client
	c, err := cache.New("redis", cache.Address("redis://127.0.0.1:6379"))
	assert.NoError(t, err)
	// helper function
	// mock the data in the redis
	mock := func(count int, prefix string) {
		for i := 0; i < count; i++ {
			err = c.Save(context.TODO(), fmt.Sprintf("%s-%d", prefix, i), "", 0)
			assert.NoError(t, err)
		}
	}
	// check after running delKeys, should no keys found
	afterCheck := func(prefix string) {
		iter, err := c.Scan(context.TODO(), prefix)
		assert.NoError(t, err)
		assert.False(t, iter.Next(context.TODO()))
	}

	{
		prefix := "mock-group-1"
		count := 100
		mock(count, prefix)
		assert.NoError(t, delKeys(context.TODO(), c, prefix))
		afterCheck(prefix)
	}

	{
		prefix := "mock-group-2"
		count := 1100
		mock(count, prefix)
		assert.NoError(t, delKeys(context.TODO(), c, prefix))
		afterCheck(prefix)
	}
}
