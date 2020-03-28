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

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/lib/errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Build query sting and pagination information into the Query model
// query string format: q=k=v,k=~v,k=[min~max],k={v1 v2 v3},k=(v1 v2 v3)
// exact match: k=v
// fuzzy match: k=~v
// range: k=[min~max]
// or list: k={v1 v2 v3}
// and list: k=(v1 v2 v3)
func Build(q string, pageNumber, pageSize int64) (*Query, error) {
	query := &Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
		Keywords:   map[string]interface{}{},
	}
	if len(q) == 0 {
		return query, nil
	}
	// try to escaped the 'q=tags%3Dnil' when to filter tags.
	if unescapedQuery, err := url.QueryUnescape(q); err == nil {
		q = unescapedQuery
	} else {
		log.Errorf("failed to unescape the query %s: %v", q, err)
	}
	params := strings.Split(q, ",")
	for _, param := range params {
		strs := strings.SplitN(param, "=", 2)
		if len(strs) != 2 || len(strs[0]) == 0 || len(strs[1]) == 0 {
			return nil, errors.New(nil).
				WithCode(errors.BadRequestCode).
				WithMessage(`the query string must contain "=" and the key/value cannot be empty`)
		}
		value, err := parsePattern(strs[1])
		if err != nil {
			return nil, errors.New(err).
				WithCode(errors.BadRequestCode).
				WithMessage("invalid query string value: %s", strs[1])
		}
		query.Keywords[strs[0]] = value
	}
	return query, nil
}

func parsePattern(value string) (interface{}, error) {
	// empty string
	if len(value) == 0 {
		return value, nil
	}
	switch value[0] {
	case '~':
		return parseFuzzyMatchValue(value)
	case '[':
		return parseRange(value)
	case '{':
		return parseOrList(value)
	case '(':
		return parseAndList(value)
	default:
		// others: exact match
		return escapeValue(value), nil
	}
}

func parseFuzzyMatchValue(value string) (*FuzzyMatchValue, error) {
	if len(value) < 2 || value[0] != '~' {
		return nil, fmt.Errorf(`fuzzy match value must start with "~" and contain at least 1 other characters`)
	}
	return &FuzzyMatchValue{
		Value: value[1:],
	}, nil
}

func parseRange(value string) (*Range, error) {
	length := len(value)
	if value[length-1] != ']' || strings.Count(value, "~") != 1 {
		return nil, fmt.Errorf(`range must start with "[", end with "]" and contains only one "~"`)
	}
	strs := strings.SplitN(value[1:length-1], "~", 2)
	min := strings.TrimSpace(strs[0])
	max := strings.TrimSpace(strs[1])
	if len(min) == 0 && len(max) == 0 {
		return nil, fmt.Errorf(`min and max at least one should be set in range'`)
	}
	r := &Range{}
	if len(min) > 0 {
		r.Min = parseValue(min)
	}
	if len(max) > 0 {
		r.Max = parseValue(max)
	}
	return r, nil
}

func parseOrList(value string) (*OrList, error) {
	values, err := parseList(value, '{')
	if err != nil {
		return nil, err
	}
	ol := &OrList{}
	for _, v := range values {
		ol.Values = append(ol.Values, v)
	}
	return ol, nil
}

func parseAndList(value string) (*AndList, error) {
	values, err := parseList(value, '(')
	if err != nil {
		return nil, err
	}
	al := &AndList{}
	for _, v := range values {
		al.Values = append(al.Values, v)
	}
	return al, nil
}

func parseList(value string, c rune) ([]interface{}, error) {
	length := len(value)
	if c == '{' && value[length-1] != '}' {
		return nil, fmt.Errorf(`or list must start with "{" and end with "}"`)
	}
	if c == '(' && value[length-1] != ')' {
		return nil, fmt.Errorf(`and list must start with "(" and end with ")"`)
	}
	var vs []interface{}
	strs := strings.Split(value[1:length-1], " ")
	for _, str := range strs {
		v := parseValue(str)
		if s, ok := v.(string); ok && len(s) == 0 {
			continue
		}
		vs = append(vs, v)
	}
	return vs, nil
}

// try to parse value as time first, then integer, and last string
func parseValue(value string) interface{} {
	value = strings.TrimSpace(value)
	// try to parse time
	time, err := time.Parse("2006-01-02T15:04:05", value)
	if err == nil {
		return time
	}
	// try to parse integer
	i, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return i
	}
	// if the value isn't time and integer, treat it as string
	return strings.Trim(value, `"'`)
}

// escape the special character
func escapeValue(value string) string {
	if len(value) > 0 && value[0] == '\\' {
		return value[1:]
	}
	return value
}
