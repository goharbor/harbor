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

	"github.com/stretchr/testify/assert"
)

func TestSend_legacy_no_starttls(t *testing.T) {
	t.Parallel()
	identity := ""
	username := "harbortestonly@gmail.com"
	password := "harborharbor"
	timeout := 60
	insecure := false
	from := "from"
	to := []string{username}
	subject := "subject"
	message := "message"

	addr := "smtp.gmail.com:465"
	tls := true
	err := Send(addr, identity, username, password,
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

func TestSend(t *testing.T) {
	t.Parallel()
	identity := ""
	username := "harbortestonly@gmail.com"
	password := "harborharbor"
	timeout := 60
	insecure := false
	from := "from"
	to := []string{username}
	subject := "subject"
	message := "message"

	addr := "smtp.gmail.com:587"
	tls := true
	err := Send(addr, identity, username, password,
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

func TestPing_legacy_no_starttls(t *testing.T) {
	t.Parallel()
	identity := ""
	username := "harbortestonly@gmail.com"
	password := "harborharbor"
	timeout := 0
	insecure := false

	addr := "smtp.gmail.com:465"
	tls := false
	err := Ping(addr, identity, username, password,
		timeout, tls, insecure)
	if err == nil {
		t.Errorf("there should be an auth error")
	} else {
		if !strings.Contains(err.Error(), "535") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestPing(t *testing.T) {
	t.Parallel()
	identity := ""
	username := "harbortestonly@gmail.com"
	password := "harborharbor"
	timeout := 0
	insecure := false

	addr := "smtp.gmail.com:587"
	tls := true
	err := Ping(addr, identity, username, password,
		timeout, tls, insecure)
	if err == nil {
		t.Errorf("there should be an auth error")
	} else {
		assert.Error(t, err, "'email_ssl' is true but \"smtp.gmail.com\" does not advertise the STARTTLS extension")
	}
}

func TestSend_exchange_invalid_auth_fails(t *testing.T) {
	t.Parallel()
	addr := "smtp.office365.com:587"
	identity := ""
	username := "someone@someorg.com"
	password := "hunter2"
	timeout := 0
	tls := true
	insecure := false
	from := "someone@someorg.com"
	to := []string{username}
	subject := "subject"
	message := "message"

	err := Send(addr, identity, username, password,
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

func TestSend_exchange(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: create a real account so we can test if sending succeeds")

	addr := "smtp.office365.com:587"
	identity := ""
	username := "someone@someorg.com"
	password := "hunter2"
	timeout := 0
	tls := true
	insecure := false
	from := "someone@someorg.com"
	to := []string{username}
	subject := "subject"
	message := "message"

	err := Send(addr, identity, username, password,
		timeout, tls, insecure, from, to,
		subject, message)
	assert.NoError(t, err)
}

func TestSend_exchange_no_tls_fails(t *testing.T) {
	t.Parallel()
	addr := "smtp.office365.com:587"
	identity := ""
	username := "someone@someorg.com"
	password := "hunter2"
	timeout := 0
	tls := false
	insecure := false
	from := "someone@someorg.com"
	to := []string{username}
	subject := "subject"
	message := "message"

	err := Send(addr, identity, username, password,
		timeout, tls, insecure, from, to,
		subject, message)
	if err == nil {
		t.Errorf("there should be an error")
	} else {
		if !strings.Contains(err.Error(), "535") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestPing_exchange(t *testing.T) {
	t.Parallel()
	addr := "smtp.office365.com:587"
	identity := ""
	username := "someone@someorg.com"
	password := "hunter2"
	timeout := 0
	tls := false
	insecure := false

	err := Ping(addr, identity, username, password,
		timeout, tls, insecure)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "535")
}
