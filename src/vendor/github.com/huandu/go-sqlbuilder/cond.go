// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"strings"
)

// Cond provides several helper methods to build conditions.
type Cond struct {
	Args *Args
}

// Equal represents "field = value".
func (c *Cond) Equal(field string, value interface{}) string {
	return fmt.Sprintf("%s = %s", Escape(field), c.Args.Add(value))
}

// E is an alias of Equal.
func (c *Cond) E(field string, value interface{}) string {
	return c.Equal(field, value)
}

// NotEqual represents "field != value".
func (c *Cond) NotEqual(field string, value interface{}) string {
	return fmt.Sprintf("%s <> %s", Escape(field), c.Args.Add(value))
}

// NE is an alias of NotEqual.
func (c *Cond) NE(field string, value interface{}) string {
	return c.NotEqual(field, value)
}

// GreaterThan represents "field > value".
func (c *Cond) GreaterThan(field string, value interface{}) string {
	return fmt.Sprintf("%s > %s", Escape(field), c.Args.Add(value))
}

// G is an alias of GreaterThan.
func (c *Cond) G(field string, value interface{}) string {
	return c.GreaterThan(field, value)
}

// GreaterEqualThan represents "field >= value".
func (c *Cond) GreaterEqualThan(field string, value interface{}) string {
	return fmt.Sprintf("%s >= %s", Escape(field), c.Args.Add(value))
}

// GE is an alias of GreaterEqualThan.
func (c *Cond) GE(field string, value interface{}) string {
	return c.GreaterEqualThan(field, value)
}

// LessThan represents "field < value".
func (c *Cond) LessThan(field string, value interface{}) string {
	return fmt.Sprintf("%s < %s", Escape(field), c.Args.Add(value))
}

// L is an alias of LessThan.
func (c *Cond) L(field string, value interface{}) string {
	return c.LessThan(field, value)
}

// LessEqualThan represents "field <= value".
func (c *Cond) LessEqualThan(field string, value interface{}) string {
	return fmt.Sprintf("%s <= %s", Escape(field), c.Args.Add(value))
}

// LE is an alias of LessEqualThan.
func (c *Cond) LE(field string, value interface{}) string {
	return c.LessEqualThan(field, value)
}

// In represents "field IN (value...)".
func (c *Cond) In(field string, value ...interface{}) string {
	vs := make([]string, 0, len(value))

	for _, v := range value {
		vs = append(vs, c.Args.Add(v))
	}

	return fmt.Sprintf("%s IN (%s)", Escape(field), strings.Join(vs, ", "))
}

// NotIn represents "field NOT IN (value...)".
func (c *Cond) NotIn(field string, value ...interface{}) string {
	vs := make([]string, 0, len(value))

	for _, v := range value {
		vs = append(vs, c.Args.Add(v))
	}

	return fmt.Sprintf("%s NOT IN (%s)", Escape(field), strings.Join(vs, ", "))
}

// Like represents "field LIKE value".
func (c *Cond) Like(field string, value interface{}) string {
	return fmt.Sprintf("%s LIKE %s", Escape(field), c.Args.Add(value))
}

// NotLike represents "field NOT LIKE value".
func (c *Cond) NotLike(field string, value interface{}) string {
	return fmt.Sprintf("%s NOT LIKE %s", Escape(field), c.Args.Add(value))
}

// IsNull represents "field IS NULL".
func (c *Cond) IsNull(field string) string {
	return fmt.Sprintf("%s IS NULL", Escape(field))
}

// IsNotNull represents "field IS NOT NULL".
func (c *Cond) IsNotNull(field string) string {
	return fmt.Sprintf("%s IS NOT NULL", Escape(field))
}

// Between represents "field BETWEEN lower AND upper".
func (c *Cond) Between(field string, lower, upper interface{}) string {
	return fmt.Sprintf("%s BETWEEN %s AND %s", Escape(field), c.Args.Add(lower), c.Args.Add(upper))
}

// NotBetween represents "field NOT BETWEEN lower AND upper".
func (c *Cond) NotBetween(field string, lower, upper interface{}) string {
	return fmt.Sprintf("%s NOT BETWEEN %s AND %s", Escape(field), c.Args.Add(lower), c.Args.Add(upper))
}

// Or represents OR logic like "expr1 OR expr2 OR expr3".
func (c *Cond) Or(orExpr ...string) string {
	return fmt.Sprintf("(%s)", strings.Join(orExpr, " OR "))
}

// And represents AND logic like "expr1 AND expr2 AND expr3".
func (c *Cond) And(andExpr ...string) string {
	return fmt.Sprintf("(%s)", strings.Join(andExpr, " AND "))
}

// Var returns a placeholder for value.
func (c *Cond) Var(value interface{}) string {
	return c.Args.Add(value)
}
