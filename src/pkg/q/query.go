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
}

// New returns Query with keywords
func New(kw KeyWords) *Query {
	return &Query{Keywords: kw}
}

// Copy the specified query object
func Copy(query *Query) *Query {
	if query == nil {
		return nil
	}
	q := &Query{
		PageNumber: query.PageNumber,
		PageSize:   query.PageSize,
		Keywords:   map[string]interface{}{},
	}
	for key, value := range query.Keywords {
		q.Keywords[key] = value
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
	Value interface{}
}

// RawValue indicates that the column checking shouldn't be applied to this query item
// when build the ORM query setter. It can be used to support any query filters of Beego
// ORM. e.g. &q.Query{Keywords:map[string]interface{}{"name__startswith": "slene""}}
type RawValue struct {
	Value interface{}
}
