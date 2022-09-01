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
	"context"
	"encoding/hex"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
	"github.com/google/uuid"
)

type CommittedKey struct{}

// HasCommittedKey checks whether exist committed key in context.
func HasCommittedKey(ctx context.Context) bool {
	if value := ctx.Value(CommittedKey{}); value != nil {
		return true
	}

	return false
}

// ormerTx transaction which support savepoint
type ormerTx struct {
	orm.Ormer
	txOrmer   orm.TxOrmer
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
	txOrmer, err := o.Ormer.Begin()
	if err != nil {
		return o.createSavepoint()
	}

	o.txOrmer = txOrmer
	return err
}

func (o *ormerTx) Commit() error {
	if o.savepointMode() {
		return o.releaseSavepoint()
	}

	return o.txOrmer.Commit()
}

func (o *ormerTx) Rollback() error {
	if o.savepointMode() {
		return o.rollbackToSavepoint()
	}

	return o.txOrmer.Rollback()
}
