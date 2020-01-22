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

package email

import (
	"strings"
	"testing"
	//	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	addr := "smtp.gmail.com:465"
	identity := ""
	username := "harbortestonly@gmail.com"
	password := "harborharbor"
	timeout := 60
	tls := true
	insecure := false
	from := "from"
	to := []string{username}
	subject := "subject"
	message := "message"

	// tls connection
	tls = true
	err := Send(addr, identity, username, password,
		timeout, tls, insecure, from, to,
		subject, message)
	// bypass the check due to securty policy change on gmail
	// TODO
	// assert.Nil(t, err)

	/*not work on travis
	// non-tls connection
	addr = "smtp.gmail.com:25"
	tls = false
	err = Send(addr, identity, username, password,
		timeout, tls, insecure, from, to,
		subject, message)
	assert.Nil(t, err)
	*/

	// invalid username/password
	username = "invalid_username"
	err = Send(addr, identity, username, password,
		timeout, tls, insecure, from, to,
		subject, message)
	if err == nil {
		t.Errorf("there should be an auth error")
	} else {
		if !strings.Contains(err.Error(), "535") {
			t.Errorf("unexpected error: %v", err)
		}
	}

}

func TestPing(t *testing.T) {
	addr := "smtp.gmail.com:465"
	identity := ""
	username := "harbortestonly@gmail.com"
	password := "harborharbor"
	timeout := 0
	tls := true
	insecure := false

	// tls connection
	err := Ping(addr, identity, username, password,
		timeout, tls, insecure)
	// bypass the check due to securty policy change on gmail
	// TODO
	// assert.Nil(t, err)

	/*not work on travis
	// non-tls connection
	addr = "smtp.gmail.com:25"
	tls = false
	err = Ping(addr, identity, username, password,
		timeout, tls, insecure)
	assert.Nil(t, err)
	*/

	// invalid username/password
	username = "invalid_username"
	err = Ping(addr, identity, username, password,
		timeout, tls, insecure)
	if err == nil {
		t.Errorf("there should be an auth error")
	} else {
		if !strings.Contains(err.Error(), "535") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestEmailNoUsernameStillOk(t *testing.T) {
	host := "smtp.gmail.com"
	identity := ""
	username := ""
	password := ""

	a, err := handleAuth(host, identity, username, password, "CRAM-MD5")
	if err != nil {
		t.Errorf("there should be no error")
	}
	if a != nil {
		t.Errorf("no auth method should be returned")
	}
}
