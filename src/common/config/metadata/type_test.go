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

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntType_validate(t *testing.T) {
	test := &IntType{}
	assert.NotNil(t, test.validate("sample"))
	assert.Nil(t, test.validate("1000"))

}

func TestIntType_get(t *testing.T) {
	test := &IntType{}
	result, _ := test.get("1000")
	assert.IsType(t, result, 1000)
}

func TestStringType_get(t *testing.T) {
	test := &StringType{}
	result, _ := test.get("1000")
	assert.IsType(t, result, "sample")
}

func TestStringType_validate(t *testing.T) {
	test := &StringType{}
	assert.Nil(t, test.validate("sample"))
}

func TestLdapScopeType_validate(t *testing.T) {
	test := &LdapScopeType{}
	assert.NotNil(t, test.validate("3"))
	assert.Nil(t, test.validate("2"))
}

func TestInt64Type_validate(t *testing.T) {
	test := &Int64Type{}
	assert.NotNil(t, test.validate("sample"))
	assert.Nil(t, test.validate("1000"))
}

func TestInt64Type_get(t *testing.T) {
	test := &Int64Type{}
	result, _ := test.get("32")
	assert.Equal(t, result, int64(32))
}

func TestBoolType_validate(t *testing.T) {
	test := &BoolType{}
	assert.NotNil(t, test.validate("sample"))
	assert.Nil(t, test.validate("True"))
}

func TestBoolType_get(t *testing.T) {
	test := &BoolType{}
	result, _ := test.get("true")
	assert.Equal(t, result, true)
	result, _ = test.get("false")
	assert.Equal(t, result, false)
}

func TestPasswordType_validate(t *testing.T) {
	test := &PasswordType{}
	assert.Nil(t, test.validate("zhu88jie"))
}

func TestPasswordType_get(t *testing.T) {
	test := &PasswordType{}
	assert.Nil(t, test.validate("zhu88jie"))
}

func TestMapType_validate(t *testing.T) {
	test := &MapType{}
	assert.Nil(t, test.validate(`{"sample":"abc", "another":"welcome"}`))
	assert.NotNil(t, test.validate(`{"sample":"abc", "another":"welcome"`))
}

func TestMapType_get(t *testing.T) {
	test := &MapType{}
	result, _ := test.get(`{"sample":"abc", "another":"welcome"}`)
	assert.Equal(t, map[string]interface{}{"sample": "abc", "another": "welcome"}, result)
}

func Test_parseInt64(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{"1", args{"1"}, int64(1), false},
		{"1.0", args{"1.0"}, int64(1), false},
		{"1.1", args{"1.1"}, int64(0), true},
		{"1E2", args{"1E2"}, int64(100), false},
		{"1.073741824e+11", args{"1.073741824e+11"}, int64(107374182400), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt64(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}
