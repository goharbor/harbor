//go:build go1.8
// +build go1.8

package middleware

import "net/url"

func pathUnescape(path string) (string, error) {
	return url.PathUnescape(path)
}
