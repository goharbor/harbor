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

package uaa

import (
	"fmt"
	"golang.org/x/oauth2"
)

const fakeToken = "The Fake Token"

// FakeClient is for test only
type FakeClient struct {
	Username string
	Password string
}

// PasswordAuth ...
func (fc *FakeClient) PasswordAuth(username, password string) (*oauth2.Token, error) {
	if username == fc.Username && password == fc.Password {
		return &oauth2.Token{AccessToken: fakeToken}, nil
	}
	return nil, fmt.Errorf("invalide username and password")
}

// GetUserInfo ...
func (fc *FakeClient) GetUserInfo(token string) (*UserInfo, error) {
	if token != fakeToken {
		return nil, fmt.Errorf("unexpected token: %s, expected: %s", token, fakeToken)
	}
	info := &UserInfo{
		Name:  "fakeName",
		Email: "fake@fake.com",
	}
	return info, nil
}

// UpdateConfig ...
func (fc *FakeClient) UpdateConfig(cfg *ClientConfig) error {
	return nil
}

// SearchUser ...
func (fc *FakeClient) SearchUser(name string) ([]*SearchUserEntry, error) {
	res := []*SearchUserEntry{}
	entryOne := &SearchUserEntry{
		ExtID:    "some-external-id-1",
		ID:       "u-0001",
		UserName: "one",
		Emails: []SearchUserEmailEntry{{
			Primary: false,
			Value:   "one@email.com",
		}},
	}
	entryTwoA := &SearchUserEntry{
		ExtID:    "some-external-id-2-a",
		ID:       "u-0002a",
		UserName: "two",
		Emails: []SearchUserEmailEntry{{
			Primary: false,
			Value:   "two@email.com",
		}},
	}
	entryTwoB := &SearchUserEntry{
		ExtID:    "some-external-id-2-b",
		ID:       "u-0002b",
		UserName: "two",
	}
	if name == "one" {
		res = append(res, entryOne)
	} else if name == "two" {
		res = append(res, entryTwoA)
		res = append(res, entryTwoB)
	} else if name == "error" {
		return res, fmt.Errorf("some error")
	}
	return res, nil
}
