package semver

import (
	"reflect"
	"testing"
)

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		in  string
		f   cfunc
		v   string
		err bool
	}{
		{">= 1.2", constraintGreaterThanEqual, "1.2.0", false},
		{"1.0", constraintTildeOrEqual, "1.0.0", false},
		{"foo", nil, "", true},
		{"<= 1.2", constraintLessThanEqual, "1.2.0", false},
		{"=< 1.2", constraintLessThanEqual, "1.2.0", false},
		{"=> 1.2", constraintGreaterThanEqual, "1.2.0", false},
		{"v1.2", constraintTildeOrEqual, "1.2.0", false},
		{"=1.5", constraintTildeOrEqual, "1.5.0", false},
		{"> 1.3", constraintGreaterThan, "1.3.0", false},
		{"< 1.4.1", constraintLessThan, "1.4.1", false},
	}

	for _, tc := range tests {
		c, err := parseConstraint(tc.in)
		if tc.err && err == nil {
			t.Errorf("Expected error for %s didn't occur", tc.in)
		} else if !tc.err && err != nil {
			t.Errorf("Unexpected error for %s", tc.in)
		}

		// If an error was expected continue the loop and don't try the other
		// tests as they will cause errors.
		if tc.err {
			continue
		}

		if tc.v != c.con.String() {
			t.Errorf("Incorrect version found on %s", tc.in)
		}

		f1 := reflect.ValueOf(tc.f)
		f2 := reflect.ValueOf(c.function)
		if f1 != f2 {
			t.Errorf("Wrong constraint found for %s", tc.in)
		}
	}
}

func TestConstraintCheck(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		check      bool
	}{
		{"= 2.0", "1.2.3", false},
		{"= 2.0", "2.0.0", true},
		{"4.1", "4.1.0", true},
		{"!=4.1", "4.1.0", false},
		{"!=4.1", "5.1.0", true},
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},
		{">=1.1", "4.1.0", true},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "1.1.0", true},
		{"<=1.1", "1.1.1", false},
		{">0", "0.0.1-alpha", true},
		{">=0", "0.0.1-alpha", true},
		{">0", "0", false},
		{">=0", "0", true},
		{"=0", "1", false},
	}

	for _, tc := range tests {
		c, err := parseConstraint(tc.constraint)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		v, err := NewVersion(tc.version)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		a := c.check(v)
		if a != tc.check {
			t.Errorf("Constraint %q failing with %q", tc.constraint, tc.version)
		}
	}
}

func TestNewConstraint(t *testing.T) {
	tests := []struct {
		input string
		ors   int
		count int
		err   bool
	}{
		{">= 1.1", 1, 1, false},
		{"2.0", 1, 1, false},
		{"v2.3.5-20161202202307-sha.e8fc5e5", 1, 1, false},
		{">= bar", 0, 0, true},
		{">= 1.2.3, < 2.0", 1, 2, false},
		{">= 1.2.3, < 2.0 || => 3.0, < 4", 2, 2, false},

		// The 3 - 4 should be broken into 2 by the range rewriting
		{"3 - 4 || => 3.0, < 4", 2, 2, false},
	}

	for _, tc := range tests {
		v, err := NewConstraint(tc.input)
		if tc.err && err == nil {
			t.Errorf("expected but did not get error for: %s", tc.input)
			continue
		} else if !tc.err && err != nil {
			t.Errorf("unexpectederror for input %s: %s", tc.input, err)
			continue
		}
		if tc.err {
			continue
		}

		l := len(v.constraints)
		if tc.ors != l {
			t.Errorf("Expected %s to have %d ORs but got %d",
				tc.input, tc.ors, l)
		}

		l = len(v.constraints[0])
		if tc.count != l {
			t.Errorf("Expected %s to have %d constraints but got %d",
				tc.input, tc.count, l)
		}
	}
}

func TestConstraintsCheck(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		check      bool
	}{
		{"*", "1.2.3", true},
		{"~0.0.0", "1.2.3", true},
		{"0.x.x", "1.2.3", false},
		{"0.0.x", "1.2.3", false},
		{"0.0.0", "1.2.3", false},
		{"*", "1.2.3", true},
		{"^0.0.0", "1.2.3", false},
		{"= 2.0", "1.2.3", false},
		{"= 2.0", "2.0.0", true},
		{"4.1", "4.1.0", true},
		{"4.1.x", "4.1.3", true},
		{"1.x", "1.4", true},
		{"!=4.1", "4.1.0", false},
		{"!=4.1-alpha", "4.1.0-alpha", false},
		{"!=4.1-alpha", "4.1.0", true},
		{"!=4.1", "5.1.0", true},
		{"!=4.x", "5.1.0", true},
		{"!=4.x", "4.1.0", false},
		{"!=4.1.x", "4.2.0", true},
		{"!=4.2.x", "4.2.3", false},
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},
		{"<1.x", "1.1.1", true},
		{"<1.x", "2.1.1", false},
		{"<1.1.x", "1.2.1", false},
		{"<1.1.x", "1.1.500", true},
		{"<1.2.x", "1.1.1", true},
		{">=1.1", "4.1.0", true},
		{">=1.1", "4.1.0-beta", false},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "0.1.0-alpha", false},
		{"<=1.1-a", "0.1.0-alpha", true},
		{"<=1.1", "1.1.0", true},
		{"<=1.x", "1.1.0", true},
		{"<=2.x", "3.1.0", false},
		{"<=1.1", "1.1.1", false},
		{"<=1.1.x", "1.2.500", false},
		{">1.1, <2", "1.1.1", true},
		{">1.1, <3", "4.3.2", false},
		{">=1.1, <2, !=1.2.3", "1.2.3", false},
		{">=1.1, <2, !=1.2.3 || > 3", "3.1.2", true},
		{">=1.1, <2, !=1.2.3 || >= 3", "3.0.0", true},
		{">=1.1, <2, !=1.2.3 || > 3", "3.0.0", false},
		{">=1.1, <2, !=1.2.3 || > 3", "1.2.3", false},
		{"1.1 - 2", "1.1.1", true},
		{"1.1-3", "4.3.2", false},
		{"^1.1", "1.1.1", true},
		{"^1.1", "4.3.2", false},
		{"^1.x", "1.1.1", true},
		{"^2.x", "1.1.1", false},
		{"^1.x", "2.1.1", false},
		{"^1.x", "1.1.1-beta1", false},
		{"^1.1.2-alpha", "1.2.1-beta1", true},
		{"^1.2.x-alpha", "1.1.1-beta1", false},
		{"~*", "2.1.1", true},
		{"~1", "2.1.1", false},
		{"~1", "1.3.5", true},
		{"~1", "1.4", true},
		{"~1.x", "2.1.1", false},
		{"~1.x", "1.3.5", true},
		{"~1.x", "1.4", true},
		{"~1.1", "1.1.1", true},
		{"~1.1", "1.1.1-alpha", false},
		{"~1.1-alpha", "1.1.1-beta", true},
		{"~1.1.1-beta", "1.1.1-alpha", false},
		{"~1.1.1-beta", "1.1.1", true},
		{"~1.2.3", "1.2.5", true},
		{"~1.2.3", "1.2.2", false},
		{"~1.2.3", "1.3.2", false},
		{"~1.1", "1.2.3", false},
		{"~1.3", "2.4.5", false},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		v, err := NewVersion(tc.version)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		a := c.Check(v)
		if a != tc.check {
			t.Errorf("Constraint '%s' failing with '%s'", tc.constraint, tc.version)
		}
	}
}

func TestRewriteRange(t *testing.T) {
	tests := []struct {
		c  string
		nc string
	}{
		{"2 - 3", ">= 2, <= 3"},
		{"2 - 3, 2 - 3", ">= 2, <= 3,>= 2, <= 3"},
		{"2 - 3, 4.0.0 - 5.1", ">= 2, <= 3,>= 4.0.0, <= 5.1"},
	}

	for _, tc := range tests {
		o := rewriteRange(tc.c)

		if o != tc.nc {
			t.Errorf("Range %s rewritten incorrectly as '%s'", tc.c, o)
		}
	}
}

func TestIsX(t *testing.T) {
	tests := []struct {
		t string
		c bool
	}{
		{"A", false},
		{"%", false},
		{"X", true},
		{"x", true},
		{"*", true},
	}

	for _, tc := range tests {
		a := isX(tc.t)
		if a != tc.c {
			t.Errorf("Function isX error on %s", tc.t)
		}
	}
}

func TestConstraintsValidate(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		check      bool
	}{
		{"*", "1.2.3", true},
		{"~0.0.0", "1.2.3", true},
		{"= 2.0", "1.2.3", false},
		{"= 2.0", "2.0.0", true},
		{"4.1", "4.1.0", true},
		{"4.1.x", "4.1.3", true},
		{"1.x", "1.4", true},
		{"!=4.1", "4.1.0", false},
		{"!=4.1", "5.1.0", true},
		{"!=4.x", "5.1.0", true},
		{"!=4.x", "4.1.0", false},
		{"!=4.1.x", "4.2.0", true},
		{"!=4.2.x", "4.2.3", false},
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},
		{"<1.x", "1.1.1", true},
		{"<1.x", "2.1.1", false},
		{"<1.1.x", "1.2.1", false},
		{"<1.1.x", "1.1.500", true},
		{"<1.2.x", "1.1.1", true},
		{">=1.1", "4.1.0", true},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "1.1.0", true},
		{"<=1.x", "1.1.0", true},
		{"<=2.x", "3.1.0", false},
		{"<=1.1", "1.1.1", false},
		{"<=1.1.x", "1.2.500", false},
		{">1.1, <2", "1.1.1", true},
		{">1.1, <3", "4.3.2", false},
		{">=1.1, <2, !=1.2.3", "1.2.3", false},
		{">=1.1, <2, !=1.2.3 || > 3", "3.1.2", true},
		{">=1.1, <2, !=1.2.3 || >= 3", "3.0.0", true},
		{">=1.1, <2, !=1.2.3 || > 3", "3.0.0", false},
		{">=1.1, <2, !=1.2.3 || > 3", "1.2.3", false},
		{"1.1 - 2", "1.1.1", true},
		{"1.1-3", "4.3.2", false},
		{"^1.1", "1.1.1", true},
		{"^1.1", "1.1.1-alpha", false},
		{"^1.1.1-alpha", "1.1.1-beta", true},
		{"^1.1.1-beta", "1.1.1-alpha", false},
		{"^1.1", "4.3.2", false},
		{"^1.x", "1.1.1", true},
		{"^2.x", "1.1.1", false},
		{"^1.x", "2.1.1", false},
		{"~*", "2.1.1", true},
		{"~1", "2.1.1", false},
		{"~1", "1.3.5", true},
		{"~1", "1.3.5-beta", false},
		{"~1.x", "2.1.1", false},
		{"~1.x", "1.3.5", true},
		{"~1.x", "1.3.5-beta", false},
		{"~1.3.6-alpha", "1.3.5-beta", false},
		{"~1.3.5-alpha", "1.3.5-beta", true},
		{"~1.3.5-beta", "1.3.5-alpha", false},
		{"~1.x", "1.4", true},
		{"~1.1", "1.1.1", true},
		{"~1.2.3", "1.2.5", true},
		{"~1.2.3", "1.2.2", false},
		{"~1.2.3", "1.3.2", false},
		{"~1.1", "1.2.3", false},
		{"~1.3", "2.4.5", false},
	}

	for _, tc := range tests {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		v, err := NewVersion(tc.version)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		a, msgs := c.Validate(v)
		if a != tc.check {
			t.Errorf("Constraint '%s' failing with '%s'", tc.constraint, tc.version)
		} else if !a && len(msgs) == 0 {
			t.Errorf("%q failed with %q but no errors returned", tc.constraint, tc.version)
		}

		// if a == false {
		// 	for _, m := range msgs {
		// 		t.Errorf("%s", m)
		// 	}
		// }
	}

	v, err := NewVersion("1.2.3")
	if err != nil {
		t.Errorf("err: %s", err)
	}

	c, err := NewConstraint("!= 1.2.5, ^2, <= 1.1.x")
	if err != nil {
		t.Errorf("err: %s", err)
	}

	_, msgs := c.Validate(v)
	if len(msgs) != 2 {
		t.Error("Invalid number of validations found")
	}
	e := msgs[0].Error()
	if e != "1.2.3 does not have same major version as 2" {
		t.Error("Did not get expected message: 1.2.3 does not have same major version as 2")
	}
	e = msgs[1].Error()
	if e != "1.2.3 is greater than 1.1.x" {
		t.Error("Did not get expected message: 1.2.3 is greater than 1.1.x")
	}

	tests2 := []struct {
		constraint, version, msg string
	}{
		{"= 2.0", "1.2.3", "1.2.3 is not equal to 2.0"},
		{"!=4.1", "4.1.0", "4.1.0 is equal to 4.1"},
		{"!=4.x", "4.1.0", "4.1.0 is equal to 4.x"},
		{"!=4.2.x", "4.2.3", "4.2.3 is equal to 4.2.x"},
		{">1.1", "1.1.0", "1.1.0 is less than or equal to 1.1"},
		{"<1.1", "1.1.0", "1.1.0 is greater than or equal to 1.1"},
		{"<1.1", "1.1.1", "1.1.1 is greater than or equal to 1.1"},
		{"<1.x", "2.1.1", "2.1.1 is greater than or equal to 1.x"},
		{"<1.1.x", "1.2.1", "1.2.1 is greater than or equal to 1.1.x"},
		{">=1.1", "0.0.9", "0.0.9 is less than 1.1"},
		{"<=2.x", "3.1.0", "3.1.0 is greater than 2.x"},
		{"<=1.1", "1.1.1", "1.1.1 is greater than 1.1"},
		{"<=1.1.x", "1.2.500", "1.2.500 is greater than 1.1.x"},
		{">1.1, <3", "4.3.2", "4.3.2 is greater than or equal to 3"},
		{">=1.1, <2, !=1.2.3", "1.2.3", "1.2.3 is equal to 1.2.3"},
		{">=1.1, <2, !=1.2.3 || > 3", "3.0.0", "3.0.0 is greater than or equal to 2"},
		{">=1.1, <2, !=1.2.3 || > 3", "1.2.3", "1.2.3 is equal to 1.2.3"},
		{"1.1 - 3", "4.3.2", "4.3.2 is greater than 3"},
		{"^1.1", "4.3.2", "4.3.2 does not have same major version as 1.1"},
		{"^2.x", "1.1.1", "1.1.1 does not have same major version as 2.x"},
		{"^1.x", "2.1.1", "2.1.1 does not have same major version as 1.x"},
		{"~1", "2.1.2", "2.1.2 does not have same major and minor version as 1"},
		{"~1.x", "2.1.1", "2.1.1 does not have same major and minor version as 1.x"},
		{"~1.2.3", "1.2.2", "1.2.2 does not have same major and minor version as 1.2.3"},
		{"~1.2.3", "1.3.2", "1.3.2 does not have same major and minor version as 1.2.3"},
		{"~1.1", "1.2.3", "1.2.3 does not have same major and minor version as 1.1"},
		{"~1.3", "2.4.5", "2.4.5 does not have same major and minor version as 1.3"},
	}

	for _, tc := range tests2 {
		c, err := NewConstraint(tc.constraint)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		v, err := NewVersion(tc.version)
		if err != nil {
			t.Errorf("err: %s", err)
			continue
		}

		_, msgs := c.Validate(v)
		e := msgs[0].Error()
		if e != tc.msg {
			t.Errorf("Did not get expected message %q: %s", tc.msg, e)
		}
	}
}
