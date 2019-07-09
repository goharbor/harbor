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

package retention

import "github.com/goharbor/harbor/src/pkg/retention/res"

// Client is designed to access core service to get required infos
type Client interface {
	// Get the tag candidates under the repository
	//
	//  Arguments:
	//    repo string : name of the repository with namespace
	//
	//  Returns:
	//    []*res.Candidate : candidates returned
	//    error            : common error if any errors occurred
	GetCandidates(repo string) ([]*res.Candidate, error)

	// Delete the specified candidate
	//
	//  Arguments:
	//    candidate *res.Candidate : the deleting candidate
	//
	//  Returns:
	//    error : common error if any errors occurred
	Delete(candidate *res.Candidate) error
}

// New basic client
func New() Client {
	return &basicClient{}
}

// basicClient is a default
type basicClient struct{}

// GetCandidates gets the tag candidates under the repository
func (bc *basicClient) GetCandidates(repo string) ([]*res.Candidate, error) {
	return nil, nil
}

// Deletes the specified candidate
func (bc *basicClient) Delete(candidate *res.Candidate) error {
	return nil
}
