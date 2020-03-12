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

package blob

import (
	"github.com/stretchr/testify/mock"
)

// FakeFetcher is a fake blob fetcher that implement the src/api/artifact/abstractor/blob.Fetcher interface
type FakeFetcher struct {
	mock.Mock
}

// FetchManifest ...
func (f *FakeFetcher) FetchManifest(repoFullName, digest string) (string, []byte, error) {
	args := f.Called(mock.Anything)
	return args.String(0), args.Get(1).([]byte), args.Error(2)
}

// FetchLayer ...
func (f *FakeFetcher) FetchLayer(repoFullName, digest string) (content []byte, err error) {
	args := f.Called(mock.Anything)
	return args.Get(0).([]byte), args.Error(1)
}
