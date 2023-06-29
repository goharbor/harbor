package gc

import (
	"github.com/goharbor/harbor/src/lib/errors"
	"testing"

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
