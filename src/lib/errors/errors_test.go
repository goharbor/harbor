// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestErrCode(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nil", args{nil}, ""},
		{"general err", args{errors.New("general err")}, GeneralCode},
		{"code in err", args{&Error{Code: "code in err"}}, "code in err"},
		{"code in cause", args{&Error{Cause: &Error{Code: "code in cause"}}}, "code in cause"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrCode(tt.args.err); got != tt.want {
				t.Errorf("ErrCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

const (
	caller1Stack = "errors.caller1"
	caller2Stack = "errors.caller2"
	caller3Stack = "errors.caller3"
)

func caller1() error {
	err := caller2()
	return err
}

func caller2() error {
	err := caller3()
	return err
}

func caller3() error {
	err := caller4()
	return New(err).WithMessage("it's caller 3.")
}

func caller4() error {
	return errors.New("it's caller 4")
}

type ErrorTestSuite struct {
	suite.Suite
}

func (suite *ErrorTestSuite) TestNewCompatibleWithStdlib() {
	err1 := New("oops")
	err2 := errors.New("oops")

	suite.Equal(err2.Error(), err1.Error())
}

func (suite *ErrorTestSuite) TestStackTrace() {
	err := caller1()
	suite.Contains(err.(*Error).StackTrace(), caller1Stack)
	suite.Contains(err.(*Error).StackTrace(), caller2Stack)
	suite.Contains(err.(*Error).StackTrace(), caller3Stack)
	suite.Contains(err.Error(), "it's caller 3.")
	suite.Contains(err.Error(), "it's caller 4")
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, &ErrorTestSuite{})
}
