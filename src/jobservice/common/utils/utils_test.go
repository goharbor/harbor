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

// Package utils provides reusable and sharable utilities for other packages and components.
package utils

import (
	"os"
	"testing"

	"github.com/gocraft/work"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UtilsTestSuite tests the utils package
type UtilsTestSuite struct {
	suite.Suite
}

// TestUtilsTestSuite is suite entry for 'go test'
func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

// TestMakeIdentifier tests MakeIdentifier
func (suite *UtilsTestSuite) TestMakeIdentifier() {
	identifierX := MakeIdentifier()
	identifierY := MakeIdentifier()
	assert.Equal(suite.T(), 24, len(identifierX), "identifierX length should be 24")
	assert.Equal(suite.T(), 24, len(identifierY), "identifierY length should be 24")
	assert.NotEqual(suite.T(), identifierX, identifierY, "identifierX and identifierY should not be equal")
}

// TestIsEmptyStr tests IsEmptyStr
func (suite *UtilsTestSuite) TestIsEmptyStr() {
	assert.True(suite.T(), IsEmptyStr(""), "empty string should be empty")
	assert.False(suite.T(), IsEmptyStr("test"), "non-empty string should not be empty")
}

// TestReadEnv tests ReadEnv
func (suite *UtilsTestSuite) TestReadEnv() {
	os.Setenv("TEST_EXIST_ENV_VAR", "test")
	assert.Equal(suite.T(), "test", ReadEnv("TEST_EXIST_ENV_VAR"), "env var TEST_EXIST_ENV_VAR should return test value")
	assert.Equal(suite.T(), "", ReadEnv("TEST_NOT_EXIST_ENV_VAR"), "env var TEST_NOT_EXIST_ENV_VAR should return empty value")
	os.Unsetenv("TEST_EXIST_ENV_VAR")
}

// TestFileExists tests FileExists
func (suite *UtilsTestSuite) TestFileExists() {
	assert.True(suite.T(), FileExists("utils_test.go"), "utils_test.go should exist")
	assert.False(suite.T(), FileExists(""), "empty string should not exist")
	assert.False(suite.T(), FileExists("not_exist_file"), "not_exist_file should not exist")
}

// TestDirExists tests DirExists
func (suite *UtilsTestSuite) TestDirExists() {
	assert.True(suite.T(), DirExists("."), "current directory should exist")
	assert.False(suite.T(), DirExists("not_exist_dir"), "not_exist_dir should not exist")
	assert.False(suite.T(), DirExists(""), "empty string should not exist")
}

// TestIsVaildPort tests IsVaildPort
func (suite *UtilsTestSuite) TestIsValidPort() {
	assert.True(suite.T(), IsValidPort(80), "80 should be a valid port")
	assert.True(suite.T(), IsValidPort(65535), "65535 should be a valid port")
	assert.False(suite.T(), IsValidPort(65536), "65536 should not be a valid port")
	assert.False(suite.T(), IsValidPort(0), "0 should not be a valid port")
}

// TestIsValidURL tests IsValidURL
func (suite *UtilsTestSuite) TestIsValidURL() {
	assert.True(suite.T(), IsValidURL("https://www.google.com"), "https://www.google.com should be a valid URL")
	assert.True(suite.T(), IsValidURL("http://www.google.com/index"), "http://www.google.com/index should be a valid URL")
	assert.True(suite.T(), IsValidURL("www.google.com"), "www.google.com should be a valid URL")
	assert.False(suite.T(), IsValidURL(""), "empty string should not be a valid URL")
	assert.False(suite.T(), IsValidURL("not_a_url"), "not_a_url should not be a valid URL")
}

// TestJobSerializeAndDeSerialize tests SerializeJob and DeSerializeJob
func (suite *UtilsTestSuite) TestJobSerializeAndDeSerialize() {
	job := &work.Job{
		Name:       "test",
		ID:         "123",
		EnqueuedAt: 123,
		Args: map[string]interface{}{
			"test": "test",
		},
		Unique:   true,
		Fails:    0,
		LastErr:  "",
		FailedAt: 0,
	}
	serializedJob, err := SerializeJob(job)
	assert.Nil(suite.T(), err, "serialize job should be successful")
	assert.NotNil(suite.T(), serializedJob, "serialized job should not be nil")
	assert.NotEqual(suite.T(), "", string(serializedJob), "serialized job should be not empty")
	deSerializedJob, err := DeSerializeJob(serializedJob)
	assert.Nil(suite.T(), err, "deserialize job should be successful")
	assert.NotNil(suite.T(), deSerializedJob, "deserialized job should not be nil")
	assert.Equal(suite.T(), job, deSerializedJob, "deSerializeJob should equal to job")
}
