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

package trace

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJaegerConfigString(t *testing.T) {
	// Value receiver with password
	cfgVal := JaegerConfig{
		Endpoint:  "http://jaeger:14268",
		Username:  "admin",
		Password:  "secret_password_123",
		AgentHost: "localhost",
		AgentPort: "6831",
	}
	assert.NotContains(t, cfgVal.String(), "secret_password_123")
	assert.Contains(t, cfgVal.String(), "password: ******")

	// Pointer receiver with password
	cfgPtr := &JaegerConfig{
		Endpoint:  "http://jaeger:14268",
		Username:  "admin",
		Password:  "secret_password_123",
		AgentHost: "localhost",
		AgentPort: "6831",
	}
	assert.NotContains(t, cfgPtr.String(), "secret_password_123")
	assert.Contains(t, cfgPtr.String(), "password: ******")

	// Empty password
	cfgEmpty := JaegerConfig{
		Endpoint:  "http://jaeger:14268",
		Username:  "admin",
		Password:  "",
		AgentHost: "localhost",
		AgentPort: "6831",
	}
	assert.Contains(t, cfgEmpty.String(), "password: ,")
}

func TestConfigStringMasksJaegerPassword(t *testing.T) {
	c := Config{
		Enabled:     true,
		ServiceName: "harbor-core",
		Jaeger: JaegerConfig{
			Endpoint:  "http://jaeger:14268",
			Username:  "admin",
			Password:  "my_super_secret_password",
			AgentHost: "localhost",
			AgentPort: "6831",
		},
		Otel: OtelConfig{
			Endpoint: "http://otel:4317",
		},
	}

	// Direct String() call on value
	strVal := c.String()
	assert.NotContains(t, strVal, "my_super_secret_password")
	assert.Contains(t, strVal, "password: ******")

	// Direct String() call on pointer
	strPtr := (&c).String()
	assert.NotContains(t, strPtr, "my_super_secret_password")
	assert.Contains(t, strPtr, "password: ******")

	// Formatted via %v on value
	formattedVal := fmt.Sprintf("%v", c)
	assert.False(t, strings.Contains(formattedVal, "my_super_secret_password"), "fmt.Sprintf(%%v, c) exposed password: %s", formattedVal)
	assert.Contains(t, formattedVal, "password: ******")

	// Formatted via %v on pointer
	formattedPtr := fmt.Sprintf("%v", &c)
	assert.False(t, strings.Contains(formattedPtr, "my_super_secret_password"), "fmt.Sprintf(%%v, &c) exposed password: %s", formattedPtr)
	assert.Contains(t, formattedPtr, "password: ******")
}
