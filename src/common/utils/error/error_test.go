package error

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	cases := []struct {
		err      error
		reason   string
		expected bool
	}{
		{
			err:      errors.New(""),
			reason:   ReasonNotFound,
			expected: false,
		},
		{
			err: KnownError{
				Reason: ReasonNotFound,
			},
			reason:   ReasonNotFound,
			expected: true,
		},
		{
			err: KnownError{
				Reason: ReasonNotFound,
			},
			reason:   "Other",
			expected: false,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, Is(c.err, c.reason))
	}
}
