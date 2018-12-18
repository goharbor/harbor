package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTag(t *testing.T) {
	cases := []struct {
		tag   string
		valid bool
	}{
		{
			"v1.0",
			true,
		},
		{
			"1.0.0",
			true,
		},
		{
			"v1.0-alpha.0",
			true,
		},
		{
			"1__",
			true,
		},
		{
			"__v1.0",
			true,
		},
		{
			"_...",
			true,
		},
		{
			"_-_",
			true,
		},
		{
			"--v1.0",
			false,
		},
		{
			".0.1",
			false,
		},
		{
			"-0.1",
			false,
		},
		{
			"0.1.*",
			false,
		},
		{
			"0.1.?",
			false,
		},
	}

	for _, c := range cases {
		if c.valid {
			assert.True(t, ValidateTag(c.tag))
		} else {
			assert.False(t, ValidateTag(c.tag))
		}
	}
}

func TestValidateRepo(t *testing.T) {
	cases := []struct {
		repo  string
		valid bool
	}{
		{
			"a",
			true,
		},
		{
			"a_a",
			true,
		},
		{
			"a__a",
			true,
		},
		{
			"a-a",
			true,
		},
		{
			"a--a",
			true,
		},
		{
			"a---a",
			true,
		},
		{
			"a.a",
			true,
		},
		{
			"a/b.b",
			true,
		},
		{
			"a_a/b-b",
			true,
		},
		{
			".a",
			false,
		},
		{
			"_a",
			false,
		},
		{
			"-a",
			false,
		},
		{
			"a.",
			false,
		},
		{
			"a_",
			false,
		},
		{
			"a-",
			false,
		},
		{
			"a..a",
			false,
		},
		{
			"a___a",
			false,
		},
		{
			"a.-a",
			false,
		},
		{
			"a_-a",
			false,
		},
		{
			"a*",
			false,
		},
		{
			"A/_a",
			false,
		},
		{
			"A/.a",
			false,
		},
		{
			"Aaaa",
			false,
		},
		{
			"aaaA",
			false,
		},
	}

	for _, c := range cases {
		if c.valid {
			assert.True(t, ValidateRepo(c.repo))
		} else {
			assert.False(t, ValidateRepo(c.repo))
		}
	}
}
