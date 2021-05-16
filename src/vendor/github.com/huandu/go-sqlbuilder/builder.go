// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
)

// Builder is a general SQL builder.
// It's used by Args to create nested SQL like the `IN` expression in
// `SELECT * FROM t1 WHERE id IN (SELECT id FROM t2)`.
type Builder interface {
	Build() (sql string, args []interface{})
	BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{})
}

type compiledBuilder struct {
	args   *Args
	format string
}

var _ Builder = new(compiledBuilder)

func (cb *compiledBuilder) Build() (sql string, args []interface{}) {
	return cb.args.Compile(cb.format)
}

func (cb *compiledBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	return cb.args.CompileWithFlavor(cb.format, flavor, initialArg...)
}

type flavoredBuilder struct {
	builder Builder
	flavor  Flavor
}

func (fb *flavoredBuilder) Build() (sql string, args []interface{}) {
	return fb.builder.BuildWithFlavor(fb.flavor)
}

func (fb *flavoredBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	return fb.builder.BuildWithFlavor(flavor, initialArg...)
}

// WithFlavor creates a new Builder based on builder with a default flavor.
func WithFlavor(builder Builder, flavor Flavor) Builder {
	return &flavoredBuilder{
		builder: builder,
		flavor:  flavor,
	}
}

// Buildf creates a Builder from a format string using `fmt.Sprintf`-like syntax.
// As all arguments will be converted to a string internally, e.g. "$0",
// only `%v` and `%s` are valid.
func Buildf(format string, arg ...interface{}) Builder {
	args := &Args{
		Flavor: DefaultFlavor,
	}
	vars := make([]interface{}, 0, len(arg))

	for _, a := range arg {
		vars = append(vars, args.Add(a))
	}

	return &compiledBuilder{
		args:   args,
		format: fmt.Sprintf(Escape(format), vars...),
	}
}

// Build creates a Builder from a format string.
// The format string uses special syntax to represent arguments.
// See doc in `Args#Compile` for syntax details.
func Build(format string, arg ...interface{}) Builder {
	args := &Args{
		Flavor: DefaultFlavor,
	}

	for _, a := range arg {
		args.Add(a)
	}

	return &compiledBuilder{
		args:   args,
		format: format,
	}
}

// BuildNamed creates a Builder from a format string.
// The format string uses `${key}` to refer the value of named by key.
func BuildNamed(format string, named map[string]interface{}) Builder {
	args := &Args{
		Flavor:    DefaultFlavor,
		onlyNamed: true,
	}

	for n, v := range named {
		args.Add(Named(n, v))
	}

	return &compiledBuilder{
		args:   args,
		format: format,
	}
}
