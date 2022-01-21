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

package model

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AccessoryTestSuite struct {
	suite.Suite
}

func (suite *AccessoryTestSuite) SetupSuite() {
	Register("mock", func(data AccessoryData) Accessory {
		return nil
	})
}

func (suite *AccessoryTestSuite) TestNew() {
	{
		c, err := New("", AccessoryData{})
		suite.Nil(c)
		suite.Error(err)
	}

	{
		c, err := New("mocks", AccessoryData{})
		suite.Nil(c)
		suite.Error(err)
	}

	{
		c, err := New("mock", AccessoryData{})
		suite.Nil(c)
		suite.Nil(err)
	}
}

func (suite *AccessoryTestSuite) TestToAccessory() {
	data := []byte(`{"artifact_id":9,"creation_time":"2022-01-20T09:18:50.993Z","digest":"sha256:1234","icon":"","id":4,"size":501,"subject_artifact_id":8,"type":"signature.cosign"}`)
	_, err := ToAccessory(data)
	suite.NotNil(err)
}

func TestAccessoryTestSuite(t *testing.T) {
	suite.Run(t, new(AccessoryTestSuite))
}
