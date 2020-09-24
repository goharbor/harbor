package utils

import (
	"fmt"
	"regexp"
)

const nameComponent = `[a-z0-9]+((?:[._]|__|[-]*)[a-z0-9]+)*`

// TagRegexp is regular expression to match image tags, for example, 'v1.0'
var TagRegexp = regexp.MustCompile(`^[\w][\w.-]{0,127}$`)

// RepoRegexp is regular expression to match repo name, for example, 'busybox', 'stage/busybox'
var RepoRegexp = regexp.MustCompile(fmt.Sprintf("^%s(/%s)*$", nameComponent, nameComponent))

// ValidateTag validates whether a tag is valid.
func ValidateTag(tag string) bool {
	return TagRegexp.MatchString(tag)
}

// ValidateRepo validates whether a repo name is valid.
func ValidateRepo(repo string) bool {
	return RepoRegexp.MatchString(repo)
}
