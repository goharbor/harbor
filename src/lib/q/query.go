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

package q

// KeyWords ...
type KeyWords = map[string]interface{}

// Query parameters
type Query struct {
	// Page number
	PageNumber int64
	// Page size
	PageSize int64
	// List of key words
	Keywords KeyWords
	// Sorting
	Sorting string
}

// First make the query only fetch the first one record in the sorting order
func (q *Query) First(sorting ...string) *Query {
	q.PageNumber = 1
	q.PageSize = 1
	if len(sorting) > 0 {
		q.Sorting = sorting[0]
	}

	return q
}

// New returns Query with keywords
func New(kw KeyWords) *Query {
	return &Query{Keywords: kw}
}

// MustClone returns the clone of query when it's not nil
// or returns a new Query instance
func MustClone(query *Query) *Query {
	q := &Query{
		Keywords: map[string]interface{}{},
	}
	if query != nil {
		q.PageNumber = query.PageNumber
		q.PageSize = query.PageSize
		q.Sorting = query.Sorting
		for k, v := range query.Keywords {
			q.Keywords[k] = v
		}
	}
	return q
}

// Range query
type Range struct {
	Min interface{}
	Max interface{}
}

// AndList query
type AndList struct {
	Values []interface{}
}

// OrList query
type OrList struct {
	Values []interface{}
}

// FuzzyMatchValue query
type FuzzyMatchValue struct {
	Value string
}
