package migrate

import (
	"errors"
	"fmt"
	nurl "net/url"
	"strings"
)

// MultiError holds multiple errors.
type MultiError struct {
	Errs []error
}

// NewMultiError returns an error type holding multiple errors.
func NewMultiError(errs ...error) MultiError {
	compactErrs := make([]error, 0)
	for _, e := range errs {
		if e != nil {
			compactErrs = append(compactErrs, e)
		}
	}
	return MultiError{compactErrs}
}

// Error implements error. Mulitple errors are concatenated with 'and's.
func (m MultiError) Error() string {
	var strs = make([]string, 0)
	for _, e := range m.Errs {
		if len(e.Error()) > 0 {
			strs = append(strs, e.Error())
		}
	}
	return strings.Join(strs, " and ")
}

// suint safely converts int to uint
// see https://goo.gl/wEcqof
// see https://goo.gl/pai7Dr
func suint(n int) uint {
	if n < 0 {
		panic(fmt.Sprintf("suint(%v) expects input >= 0", n))
	}
	return uint(n)
}

var errNoScheme = errors.New("no scheme")
var errEmptyURL = errors.New("URL cannot be empty")

func sourceSchemeFromUrl(url string) (string, error) {
	u, err := schemeFromUrl(url)
	if err != nil {
		return "", fmt.Errorf("source: %v", err)
	}
	return u, nil
}

func databaseSchemeFromUrl(url string) (string, error) {
	u, err := schemeFromUrl(url)
	if err != nil {
		return "", fmt.Errorf("database: %v", err)
	}
	return u, nil
}

// schemeFromUrl returns the scheme from a URL string
func schemeFromUrl(url string) (string, error) {
	if url == "" {
		return "", errEmptyURL
	}

	u, err := nurl.Parse(url)
	if err != nil {
		return "", err
	}
	if len(u.Scheme) == 0 {
		return "", errNoScheme
	}

	return u.Scheme, nil
}

// FilterCustomQuery filters all query values starting with `x-`
func FilterCustomQuery(u *nurl.URL) *nurl.URL {
	ux := *u
	vx := make(nurl.Values)
	for k, v := range ux.Query() {
		if len(k) <= 1 || (len(k) > 1 && k[0:2] != "x-") {
			vx[k] = v
		}
	}
	ux.RawQuery = vx.Encode()
	return &ux
}
