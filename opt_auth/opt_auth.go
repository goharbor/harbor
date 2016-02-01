/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package opt_auth

import (
	"fmt"
	"os"

	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
)

type OptAuth interface {
	Validate(auth models.AuthModel) (*models.User, error)
}

var registry = make(map[string]OptAuth)

func Register(name string, optAuth OptAuth) {
	if _, dup := registry[name]; dup {
		panic(name + " already exist.")
		return
	}
	registry[name] = optAuth
}

func Login(auth models.AuthModel) (*models.User, error) {

	var authMode string = os.Getenv("AUTH_MODE")
	if authMode == "" || auth.Principal == "admin" {
		authMode = "db_auth"
	}
	beego.Debug("Current AUTH_MODE is ", authMode)

	optAuth := registry[authMode]
	if optAuth == nil {
		return nil, fmt.Errorf("Unrecognized auth_mode: %s", authMode)
	}
	return optAuth.Validate(auth)
}
