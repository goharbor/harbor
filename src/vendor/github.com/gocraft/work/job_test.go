package work

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobArgumentExtraction(t *testing.T) {
	j := Job{}
	j.setArg("str1", "bar")

	j.setArg("int1", int64(77))
	j.setArg("int2", 77)
	j.setArg("int3", uint64(77))
	j.setArg("int4", float64(77.0))

	j.setArg("bool1", true)

	j.setArg("float1", 3.14)

	//
	// Success cases:
	//
	vString := j.ArgString("str1")
	assert.Equal(t, vString, "bar")
	assert.NoError(t, j.ArgError())

	vInt64 := j.ArgInt64("int1")
	assert.EqualValues(t, vInt64, 77)
	assert.NoError(t, j.ArgError())

	vInt64 = j.ArgInt64("int2")
	assert.EqualValues(t, vInt64, 77)
	assert.NoError(t, j.ArgError())

	vInt64 = j.ArgInt64("int3")
	assert.EqualValues(t, vInt64, 77)
	assert.NoError(t, j.ArgError())

	vInt64 = j.ArgInt64("int4")
	assert.EqualValues(t, vInt64, 77)
	assert.NoError(t, j.ArgError())

	vBool := j.ArgBool("bool1")
	assert.Equal(t, vBool, true)
	assert.NoError(t, j.ArgError())

	vFloat := j.ArgFloat64("float1")
	assert.Equal(t, vFloat, 3.14)
	assert.NoError(t, j.ArgError())

	// Missing key results in error:
	vString = j.ArgString("str_missing")
	assert.Equal(t, vString, "")
	assert.Error(t, j.ArgError())
	j.argError = nil
	assert.NoError(t, j.ArgError())

	vInt64 = j.ArgInt64("int_missing")
	assert.EqualValues(t, vInt64, 0)
	assert.Error(t, j.ArgError())
	j.argError = nil
	assert.NoError(t, j.ArgError())

	vBool = j.ArgBool("bool_missing")
	assert.Equal(t, vBool, false)
	assert.Error(t, j.ArgError())
	j.argError = nil
	assert.NoError(t, j.ArgError())

	vFloat = j.ArgFloat64("float_missing")
	assert.Equal(t, vFloat, 0.0)
	assert.Error(t, j.ArgError())
	j.argError = nil
	assert.NoError(t, j.ArgError())

	// Missing string; Make sure we don't reset it with successes after
	vString = j.ArgString("str_missing")
	assert.Equal(t, vString, "")
	assert.Error(t, j.ArgError())
	_ = j.ArgString("str1")
	_ = j.ArgInt64("int1")
	_ = j.ArgBool("bool1")
	_ = j.ArgFloat64("float1")
	assert.Error(t, j.ArgError())
}

func TestJobArgumentExtractionBadString(t *testing.T) {
	var testCases = []struct {
		key  string
		val  interface{}
		good bool
	}{
		{"a", 1, false},
		{"b", false, false},
		{"c", "yay", true},
	}

	j := Job{}

	for _, tc := range testCases {
		j.setArg(tc.key, tc.val)
	}

	for _, tc := range testCases {
		r := j.ArgString(tc.key)
		err := j.ArgError()
		if tc.good {
			if err != nil {
				t.Errorf("Failed test case: %v; err = %v\n", tc, err)
			}
			if r != tc.val.(string) {
				t.Errorf("Failed test case: %v; r = %v\n", tc, r)
			}
		} else {
			if err == nil {
				t.Errorf("Failed test case: %v; but err was nil\n", tc)
			}
			if r != "" {
				t.Errorf("Failed test case: %v; but r was %v\n", tc, r)
			}
		}
		j.argError = nil
	}
}

func TestJobArgumentExtractionBadBool(t *testing.T) {
	var testCases = []struct {
		key  string
		val  interface{}
		good bool
	}{
		{"a", 1, false},
		{"b", "boo", false},
		{"c", true, true},
		{"d", false, true},
	}

	j := Job{}

	for _, tc := range testCases {
		j.setArg(tc.key, tc.val)
	}

	for _, tc := range testCases {
		r := j.ArgBool(tc.key)
		err := j.ArgError()
		if tc.good {
			if err != nil {
				t.Errorf("Failed test case: %v; err = %v\n", tc, err)
			}
			if r != tc.val.(bool) {
				t.Errorf("Failed test case: %v; r = %v\n", tc, r)
			}
		} else {
			if err == nil {
				t.Errorf("Failed test case: %v; but err was nil\n", tc)
			}
			if r != false {
				t.Errorf("Failed test case: %v; but r was %v\n", tc, r)
			}
		}
		j.argError = nil
	}
}

func TestJobArgumentExtractionBadInt(t *testing.T) {
	var testCases = []struct {
		key  string
		val  interface{}
		good bool
	}{
		{"a", "boo", false},
		{"b", true, false},
		{"c", 1.1, false},
		{"d", 19007199254740892.0, false},
		{"e", -19007199254740892.0, false},
		{"f", uint64(math.MaxInt64) + 1, false},

		{"z", 0, true},
		{"y", 9007199254740892, true},
		{"x", 9007199254740892.0, true},
		{"w", 573839921, true},
		{"v", -573839921, true},
		{"u", uint64(math.MaxInt64), true},
	}

	j := Job{}

	for _, tc := range testCases {
		j.setArg(tc.key, tc.val)
	}

	for _, tc := range testCases {
		r := j.ArgInt64(tc.key)
		err := j.ArgError()
		if tc.good {
			if err != nil {
				t.Errorf("Failed test case: %v; err = %v\n", tc, err)
			}
		} else {
			if err == nil {
				t.Errorf("Failed test case: %v; but err was nil\n", tc)
			}
			if r != 0 {
				t.Errorf("Failed test case: %v; but r was %v\n", tc, r)
			}
		}
		j.argError = nil
	}
}

func TestJobArgumentExtractionBadFloat(t *testing.T) {
	var testCases = []struct {
		key  string
		val  interface{}
		good bool
	}{
		{"a", "boo", false},
		{"b", true, false},

		{"z", 0, true},
		{"y", 9007199254740892, true},
		{"x", 9007199254740892.0, true},
		{"w", 573839921, true},
		{"v", -573839921, true},
		{"u", math.MaxFloat64, true},
		{"t", math.SmallestNonzeroFloat64, true},
	}

	j := Job{}

	for _, tc := range testCases {
		j.setArg(tc.key, tc.val)
	}

	for _, tc := range testCases {
		r := j.ArgFloat64(tc.key)
		err := j.ArgError()
		if tc.good {
			if err != nil {
				t.Errorf("Failed test case: %v; err = %v\n", tc, err)
			}
		} else {
			if err == nil {
				t.Errorf("Failed test case: %v; but err was nil\n", tc)
			}
			if r != 0 {
				t.Errorf("Failed test case: %v; but r was %v\n", tc, r)
			}
		}
		j.argError = nil
	}
}
