// Copyright 2018 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import "errors"

// Global errors for rbac defined here
var (
	ERR_NAME_NOT_FOUND    = errors.New("error: name does not exist")
	ERR_DOMAIN_PARAMETER  = errors.New("error: domain should be 1 parameter")
	ERR_NAMES12_NOT_FOUND = errors.New("error: name1 or name2 does not exist")
)
