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
	"fmt"
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
	return New(nil).WithMessage("it's caller 3.").WithCause(err)
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

func (suite *ErrorTestSuite) TestNew() {
	cause := New("root")
	suite.Equal("root", cause.Error())

	err := New(cause)
	suite.Equal("root", err.Error())

	err = New(cause.WithCause(errors.New("stdErr")))
	suite.Equal("root: stdErr", err.Error())

	errStd := errors.New("stdErr")
	err = New(errStd)
	suite.Equal("stdErr", err.Error())

	err = New(nil)
	suite.Equal("<nil>", err.Error())

	err = New("")
	suite.Equal("", err.Error())
}

func (suite *ErrorTestSuite) TestWithMessage() {
	cause := New("root")
	err := cause.WithMessage("append message").WithMessage("append message2")
	suite.Equal("append message2", err.Error())
}

func (suite *ErrorTestSuite) TestWithCause() {
	cause := errors.New("stdErr")
	err := New("root").WithCause(cause)
	suite.Equal("root: stdErr", err.Error())
}

func (suite *ErrorTestSuite) TestWithCauseMessage() {
	cause := errors.New("stdErr")
	err := New("root").WithCause(cause).WithMessage("With Message")
	suite.Equal("With Message: stdErr", err.Error())
}

func (suite *ErrorTestSuite) TestFormat() {
	err := New("empty job ID")
	suite.Equal("empty job ID", fmt.Sprintf("%s", err))
}

func (suite *ErrorTestSuite) TestWrap() {
	cause := New("root")
	cause = Wrap(cause, "err1")
	cause = Wrap(cause, "err2")
	cause = Wrap(cause, "err3")
	suite.Equal("err3: err2: err1: root", cause.Error())
}

func (suite *ErrorTestSuite) TestWrapf() {
	cause := New("root")
	cause = Wrapf(cause, "err%d", 1)
	cause = Wrapf(cause, "err%d", 2)
	cause = Wrapf(cause, "err%d", 3)
	suite.Equal("err3: err2: err1: root", cause.Error())
}

func (suite *ErrorTestSuite) TestErrof() {
	err := Errorf("it's err%d", 1)
	suite.Equal("it's err1", err.Error())
}

func (suite *ErrorTestSuite) TestErrofWithMessage() {
	err := Errorf("it's err%d", 1)
	suite.Equal("it's err1", err.Error())
}

func (suite *ErrorTestSuite) TestWrapStdErr() {
	cause := errors.New("stdErr")
	err := Wrap(cause, "wrap stdErr")
	suite.Equal("wrap stdErr: stdErr", err.Error())
}

func (suite *ErrorTestSuite) TestNilErr() {
	nilErr := New(nil)
	suite.Equal("<nil>", nilErr.Error())
}

func (suite *ErrorTestSuite) TestNilWithMessage() {
	nilErr := New(nil).WithMessage("it's a nil error")
	suite.Equal("it's a nil error", nilErr.Error())
}

func (suite *ErrorTestSuite) TestIsNotFoundErr() {
	err := New(nil).WithCode(NotFoundCode)
	suite.True(IsNotFoundErr(err))

	err = New(nil).WithCode(PreconditionCode)
	suite.False(IsNotFoundErr(err))
}

func (suite *ErrorTestSuite) TestIsConflictErrErr() {
	err := New(nil).WithCode(ConflictCode)
	suite.True(IsConflictErr(err))

	err = New(nil).WithCode(PreconditionCode)
	suite.False(IsConflictErr(err))
}

func (suite *ErrorTestSuite) TestErrCode() {
	err := New(nil).WithCode(ConflictCode)
	suite.Equal(ErrCode(err), ConflictCode)

	err = New("root")
	suite.Equal(ErrCode(err), GeneralCode)

	err = New("root")
	suite.Equal(ErrCode(err), GeneralCode)
	err.WithCode(PreconditionCode)
	suite.Equal(ErrCode(err), PreconditionCode)
	err.WithCode(DENIED)
	suite.Equal(ErrCode(err), DENIED)

	stdErr := errors.New("stdErr")
	suite.Equal(ErrCode(stdErr), GeneralCode)
	err = Wrap(stdErr, "wrap stdErr")
	suite.Equal(ErrCode(stdErr), GeneralCode)
}

func (suite *ErrorTestSuite) TestErrs() {
	err := New("root").WithCode(ConflictCode)
	suite.Equal(`{"errors":[{"code":"CONFLICT","message":"root"}]}`, NewErrs(err).Error())
}

func (suite *ErrorTestSuite) TestNotFoundError() {
	root := errors.New("something is not found")
	err := NotFoundError(root)
	suite.Equal(`resource not found: something is not found`, err.Error())

	root = errors.New("something is not found")
	err = NotFoundError(root).WithMessage("asset not found")
	suite.Equal(`asset not found: something is not found`, err.Error())
}

func (suite *ErrorTestSuite) TestIsErr() {
	stdErr := errors.New("stdErr")
	err := Wrap(stdErr, "root")
	targetErr := Wrap(err, "target err").WithCode(ConflictCode)

	suite.True(IsErr(targetErr, ConflictCode))
}

func (suite *ErrorTestSuite) TestCause() {
	// self
	root := New("root")
	suite.Equal(root, Cause(root))

	root = New("root")
	cause1 := Wrap(root, "err1")
	cause2 := Wrap(cause1, "err2")
	cause3 := Wrap(cause2, "err3")
	suite.Equal(root, Cause(cause3))
}

func (suite *ErrorTestSuite) TestCauseStd() {
	root := errors.New("stdErr")
	suite.Equal(root, Cause(root))

	root = errors.New("stdErr")
	cause1 := Wrap(root, "err1")
	cause2 := Wrap(cause1, "err2")
	cause3 := Wrap(cause2, "err3")
	suite.Equal(root, Cause(cause3))
}

func (suite *ErrorTestSuite) TestMarshalJSON() {
	stdErr := errors.New("stdErr")
	err := Wrap(stdErr, "root").WithCode(ConflictCode)
	err2 := Wrap(err, "append message").WithCode(ConflictCode)

	out, marErr := err2.MarshalJSON()
	suite.Nil(marErr)
	suite.Equal(`{"code":"CONFLICT","message":"append message: root: stdErr"}`, string(out))
}

func (suite *ErrorTestSuite) TestErrors() {
	stdErr := errors.New("stdErr")
	suite.Equal(`{"errors":[{"code":"UNKNOWN","message":"unknown: stdErr"}]}`, NewErrs(stdErr).Error())

	err := Wrap(stdErr, "root").WithCode(ConflictCode)
	err2 := Wrap(err, "append message").WithCode(ConflictCode)
	suite.Equal(`{"errors":[{"code":"CONFLICT","message":"append message: root: stdErr"}]}`, NewErrs(err2).Error())

	err = New(nil).WithCode(GeneralCode).WithMessage("internal server error")
	suite.Equal(`{"errors":[{"code":"UNKNOWN","message":"internal server error"}]}`, NewErrs(err).Error())
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, &ErrorTestSuite{})
}
