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

package orm

import (
	"github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/jackc/pgconn"
)

var (
	// ErrNoRows error from the beego orm
	ErrNoRows = orm.ErrNoRows

	// ErrOptimisticLock error when update object failed
	ErrOptimisticLock = errors.New("the object has been modified; please apply your changes to the latest version and try again")
)

// WrapNotFoundError wrap error as NotFoundError when it is orm.ErrNoRows otherwise return err
func WrapNotFoundError(err error, format string, args ...interface{}) error {
	if e := AsNotFoundError(err, format, args...); e != nil {
		return e
	}

	return err
}

// WrapConflictError wrap error as ConflictError when it is duplicate key error otherwise return err
func WrapConflictError(err error, format string, args ...interface{}) error {
	if e := AsConflictError(err, format, args...); e != nil {
		return e
	}

	return err
}

// AsNotFoundError checks whether the err is orm.ErrNoRows. If it it, wrap it
// as a src/internal/error.Error with not found error code, else return nil
func AsNotFoundError(err error, messageFormat string, args ...interface{}) *errors.Error {
	if errors.Is(err, orm.ErrNoRows) {
		e := errors.NotFoundError(nil)
		if len(messageFormat) > 0 {
			e.WithMessage(messageFormat, args...)
		}
		return e
	}
	return nil
}

// AsConflictError checks whether the err is duplicate key error. If it is, wrap it
// as a src/internal/error.Error with conflict error code, else return nil
func AsConflictError(err error, messageFormat string, args ...interface{}) *errors.Error {
	if IsDuplicateKeyError(err) {
		e := errors.New(err).
			WithCode(errors.ConflictCode).
			WithMessage(messageFormat, args...)
		return e
	}
	return nil
}

// AsForeignKeyError checks whether the err is violating foreign key constraint error. If it it, wrap it
// as a src/internal/error.Error with violating foreign key constraint error code, else return nil
func AsForeignKeyError(err error, messageFormat string, args ...interface{}) *errors.Error {
	if isViolatingForeignKeyConstraintError(err) {
		e := errors.New(err).
			WithCode(errors.ViolateForeignKeyConstraintCode).
			WithMessage(messageFormat, args...)
		return e
	}
	return nil
}

// IsDuplicateKeyError check the duplicate key error
func IsDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}

	return false
}

func isViolatingForeignKeyConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return true
	}

	return false
}
