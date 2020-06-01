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

package handler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	// with tag
	input := "library/hello-world:latest"
	repository, reference, err := parse(input)
	require.Nil(t, err)
	assert.Equal(t, "library/hello-world", repository)
	assert.Equal(t, "latest", reference)

	// with digest
	input = "library/hello-world@sha256:9572f7cdcee8591948c2963463447a53466950b3fc15a247fcad1917ca215a2f"
	repository, reference, err = parse(input)
	require.Nil(t, err)
	assert.Equal(t, "library/hello-world", repository)
	assert.Equal(t, "sha256:9572f7cdcee8591948c2963463447a53466950b3fc15a247fcad1917ca215a2f", reference)

	// invalid digest
	input = "library/hello-world@sha256:invalid_digest"
	repository, reference, err = parse(input)
	require.NotNil(t, err)

	// invalid character
	input = "library/hello-world?#:latest"
	repository, reference, err = parse(input)
	require.NotNil(t, err)

	// empty input
	input = ""
	repository, reference, err = parse(input)
	require.NotNil(t, err)
}
