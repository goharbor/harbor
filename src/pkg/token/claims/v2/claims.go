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

package v2

import (
	"crypto/subtle"
	"fmt"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/golang-jwt/jwt/v4"
)

func init() {
	jwt.MarshalSingleStringAsArray = false
}

const (
	// Issuer is the only valid issuer for jwt token sent to /v2/xxxx
	Issuer = "harbor-token-issuer"
)

// Claims represents the token claims that encapsulated in a JWT token for registry/notary resources
type Claims struct {
	jwt.RegisteredClaims
	Access []*token.ResourceActions `json:"access"`
}

// Valid checks if the issuer is harbor
func (c *Claims) Valid() error {
	if err := c.RegisteredClaims.Valid(); err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(c.Issuer), []byte(Issuer)) == 0 {
		return fmt.Errorf("invalid token issuer: %s", c.Issuer)
	}
	return nil
}
