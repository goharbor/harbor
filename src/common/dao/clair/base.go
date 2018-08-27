// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package clair

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
)

const (
	updaterLast = "updater/last"
)

var (
	ormer orm.Ormer
	once  sync.Once
)

//GetOrmer return the singleton of Ormer for clair DB.
func GetOrmer() orm.Ormer {
	once.Do(func() {
		dbInstance, err := orm.GetDB(dao.ClairDBAlias)
		if err != nil {
			panic(err)
		}
		ormer, err = orm.NewOrmWithDB("postgres", dao.ClairDBAlias, dbInstance)
		if err != nil {
			panic(err)
		}
	})
	return ormer
}

//GetLastUpdate query the table `keyvalue` in clair's DB return the value of `updater/last`
func GetLastUpdate() (int64, error) {
	var list orm.ParamsList
	num, err := GetOrmer().Raw("SELECT value from keyvalue where key=?", updaterLast).ValuesFlat(&list)
	if err != nil {
		return 0, err
	}
	if num == 1 {
		s, ok := list[0].(string)
		if !ok { // shouldn't be here.
			return 0, fmt.Errorf("The value: %v, is non-string", list[0])
		}
		res, err := strconv.ParseInt(s, 0, 64)
		if err != nil { //shouldn't be here.
			return 0, err
		}
		return res, nil
	}
	if num > 1 {
		return 0, fmt.Errorf("Multiple entries for %s in Clair DB", updaterLast)
	}
	//num is zero, it's not updated yet.
	return 0, nil
}
