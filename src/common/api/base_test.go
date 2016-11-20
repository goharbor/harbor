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
package api

import (
	"github.com/vmware/harbor/src/common/config"
	"os"
	"testing"
)

func TestGetIsInsecure(t *testing.T) {
	os.Setenv("VERIFY_REMOTE_CERT", "off")
	err := config.Reload()
	if err != nil {
		t.Errorf("Failed to load config, error: %v", err)
	}
	if !GetIsInsecure() {
		t.Errorf("GetIsInsecure() should be true when VERIFY_REMOTE_CERT is off, in fact: false")
	}
	os.Unsetenv("VERIFY_REMOTE_CERT")
}
