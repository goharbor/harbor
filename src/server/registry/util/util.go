// Copyright 2018 Project Harbor Authors
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

package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
	libhttp "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/server/router"
)

// SetLinkHeader ...
func SetLinkHeader(origURL string, n int, last string) (string, error) {
	passedURL, err := url.Parse(origURL)
	if err != nil {
		return "", err
	}
	passedURL.Fragment = ""

	v := url.Values{}
	v.Add("n", strconv.Itoa(n))
	v.Add("last", last)
	passedURL.RawQuery = v.Encode()
	urlStr := fmt.Sprintf("<%s>; rel=\"next\"", passedURL.String())

	return urlStr, nil
}

// IndexString returns the index of X in a sorts string array
// If the array is not sorted, sort it.
func IndexString(strs []string, x string) int {
	if !sort.StringsAreSorted(strs) {
		sort.Strings(strs)
	}
	i := sort.Search(len(strs), func(i int) bool { return x <= strs[i] })
	if i < len(strs) && strs[i] == x {
		return i
	}
	return -1
}

// ParseNAndLastParameters parse the n and last parameters from the query of the http request
func ParseNAndLastParameters(r *http.Request) (*int, string, error) {
	q := r.URL.Query()

	var n *int

	if q.Get("n") != "" {
		value, err := strconv.Atoi(q.Get("n"))
		if err != nil || value < 0 {
			return nil, "", errors.New(err).WithCode(errors.BadRequestCode).WithMessage("the N must be a positive int type")
		}

		n = &value
	}

	return n, q.Get("last"), nil
}

// SendListTagsResponse sends the response for list tags API
func SendListTagsResponse(w http.ResponseWriter, r *http.Request, tags []string) {
	n, last, err := ParseNAndLastParameters(r)
	if err != nil {
		libhttp.SendError(w, err)
		return
	}

	items, nextLast := pickItems(sortedAndUniqueItems(tags), n, last)

	if nextLast != "" && n != nil { // NOTE: when the nextLast is not empty the n must not be nil
		link, err := SetLinkHeader(r.URL.String(), *n, nextLast)
		if err != nil {
			libhttp.SendError(w, err)
			return
		}

		w.Header().Set("Link", link)
	}

	body := struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		Name: router.Param(r.Context(), ":splat"),
		Tags: items,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		libhttp.SendError(w, err)
	}
}

func sortedAndUniqueItems(items []string) []string {
	n := len(items)
	if n <= 1 {
		return items[0:n]
	}

	sort.Strings(items)

	j := 1
	for i := 1; i < n; i++ {
		if items[i] != items[i-1] {
			items[j] = items[i]
			j++
		}
	}

	return items[0:j]
}

// pickItems returns the first n elements which is bigger than the last from items, if the n is 0, return the empty slice
// NOTE: the items must be ordered and value of n is equal or great than 0 when n isn't nil
func pickItems(items []string, n *int, last string) ([]string, string) {
	if len(items) == 0 || (n != nil && *n == 0) {
		// no items found or request n is 0
		return []string{}, ""
	}

	if n == nil {
		l := len(items)
		n = &l
	}

	i := 0
	if last != "" {
		i = sort.Search(len(items), func(ix int) bool { return strings.Compare(items[ix], last) > 0 })
	}

	j := i + *n

	if j >= len(items) {
		j = len(items)
	}

	result := items[i:j]

	nextLast := ""
	if len(result) > 0 && items[len(items)-1] != result[len(result)-1] {
		nextLast = result[len(result)-1]
	}

	return result, nextLast
}
