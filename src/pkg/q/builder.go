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
	ierror "github.com/goharbor/harbor/src/internal/error"
	"strconv"
	"strings"
	"time"
)

// Build query sting into the Query model
// query string format: q=k=v|k=~v|k=[min,max]|k={v1,v2,v3}|k=(v1,v2,v3)|page=1|page_size=10
// exact match: k=v
// fuzzy match: k=~v
// range: k=[min,max]
// or list: k={v1,v2,v3}
// and list: k=(v1,v2,v3)
func Build(q string) (*Query, error) {
	if len(q) == 0 {
		return nil, nil
	}
	query := &Query{Keywords: map[string]interface{}{}}
	params := strings.Split(q, "|")
	for _, param := range params {
		strs := strings.SplitN(param, "=", 2)
		if len(strs) != 2 || len(strs[0]) == 0 || len(strs[1]) == 0 {
			return nil, ierror.New(nil).
				WithCode(ierror.BadRequestCode).
				WithMessage(`the query string must contain "=" and the key/value cannot be empty`)
		}
		if strs[0] == "page" {
			i, err := strconv.ParseInt(strs[1], 10, 64)
			if err != nil {
				return nil, ierror.New(nil).
					WithCode(ierror.BadRequestCode).
					WithMessage("page must be integer")
			}
			query.PageNumber = i
			continue
		}
		if strs[0] == "page_size" {
			i, err := strconv.ParseInt(strs[1], 10, 64)
			if err != nil {
				return nil, ierror.New(nil).
					WithCode(ierror.BadRequestCode).
					WithMessage("page_size must be integer")
			}
			query.PageSize = i
			continue
		}
		value, err := parsePattern(strs[1])
		if err != nil {
			return nil, ierror.New(err).
				WithCode(ierror.BadRequestCode).
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
	if len(value) < 4 || value[0] != '[' || value[len(value)-1] != ']' ||
		strings.Count(value, ",") != 1 {
		return nil, fmt.Errorf(`range must start with "[", end with "]" and contain only one ","`)
	}
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	strs := strings.SplitN(value, ",", 2)
	r := &Range{}
	if len(strs[0]) > 0 {
		r.Min = parseValue(strs[0])
	}
	if len(strs[1]) > 0 {
		r.Max = parseValue(strs[1])
	}
	return r, nil
}

func parseOrList(value string) (*OrList, error) {
	if len(value) < 2 || value[0] != '{' || value[len(value)-1] != '}' {
		return nil, fmt.Errorf(`or list must start with "{" and end with "}"`)
	}
	value = strings.TrimPrefix(value, "{")
	value = strings.TrimSuffix(value, "}")
	ol := &OrList{}
	if len(value) == 0 {
		return ol, nil
	}
	strs := strings.Split(value, ",")
	for _, str := range strs {
		ol.Values = append(ol.Values, parseValue(str))
	}
	return ol, nil
}

func parseAndList(value string) (*AndList, error) {
	if len(value) < 2 || value[0] != '(' || value[len(value)-1] != ')' {
		return nil, fmt.Errorf(`and list must start with "(" and end with ")"`)
	}
	value = strings.TrimPrefix(value, "(")
	value = strings.TrimSuffix(value, ")")
	al := &AndList{}
	if len(value) == 0 {
		return al, nil
	}
	strs := strings.Split(value, ",")
	for _, str := range strs {
		al.Values = append(al.Values, parseValue(str))
	}
	return al, nil
}

// try to parse value as time first, then integer, and last string
func parseValue(value string) interface{} {
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
