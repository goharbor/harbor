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
	"encoding/hex"
	"fmt"

	"github.com/beego/beego/orm"
	"github.com/google/uuid"
)

// ormerTx transaction which support savepoint
type ormerTx struct {
	orm.Ormer
	savepoint string
}

func (o *ormerTx) savepointMode() bool {
	return o.savepoint != ""
}

func (o *ormerTx) createSavepoint() error {
	val := uuid.New()
	o.savepoint = fmt.Sprintf("p%s", hex.EncodeToString(val[:]))

	_, err := o.Raw(fmt.Sprintf("SAVEPOINT %s", o.savepoint)).Exec()
	return err
}

func (o *ormerTx) releaseSavepoint() error {
	_, err := o.Raw(fmt.Sprintf("RELEASE SAVEPOINT %s", o.savepoint)).Exec()
	return err
}

func (o *ormerTx) rollbackToSavepoint() error {
	_, err := o.Raw(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", o.savepoint)).Exec()
	return err
}

func (o *ormerTx) Begin() error {
	err := o.Ormer.Begin()
	if err == orm.ErrTxHasBegan {
		// transaction has began for the ormer, so begin nested transaction by savepoint
		return o.createSavepoint()
	}

	return err
}

func (o *ormerTx) Commit() error {
	if o.savepointMode() {
		return o.releaseSavepoint()
	}

	return o.Ormer.Commit()
}

func (o *ormerTx) Rollback() error {
	if o.savepointMode() {
		return o.rollbackToSavepoint()
	}

	return o.Ormer.Rollback()
}
