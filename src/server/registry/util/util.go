package util

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
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
