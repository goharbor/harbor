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

type ErrorTestSuite struct {
	suite.Suite
}

func (suite *ErrorTestSuite) TestNewCompatibleWithStdlib() {
	err1 := New("oops")
	err2 := errors.New("oops")

	suite.Equal(err2.Error(), err1.Error())
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, &ErrorTestSuite{})
}
