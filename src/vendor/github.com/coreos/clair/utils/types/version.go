// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// Version represents a package version
type Version struct {
	epoch    int
	version  string
	revision string
}

var (
	// MinVersion is a special package version which is always sorted first
	MinVersion = Version{version: "#MINV#"}
	// MaxVersion is a special package version which is always sorted last
	MaxVersion = Version{version: "#MAXV#"}

	versionAllowedSymbols  = []rune{'.', '-', '+', '~', ':', '_'}
	revisionAllowedSymbols = []rune{'.', '+', '~', '_'}
)

// NewVersion function parses a string into a Version struct which can be compared
//
// The implementation is based on http://man.he.net/man5/deb-version
// on https://www.debian.org/doc/debian-policy/ch-controlfields.html#s-f-Version
//
// It uses the dpkg-1.17.25's algorithm  (lib/parsehelp.c)
func NewVersion(str string) (Version, error) {
	var version Version

	// Trim leading and trailing space
	str = strings.TrimSpace(str)

	if len(str) == 0 {
		return Version{}, errors.New("Version string is empty")
	}

	// Max/Min versions
	if str == MaxVersion.String() {
		return MaxVersion, nil
	}
	if str == MinVersion.String() {
		return MinVersion, nil
	}

	// Find epoch
	sepepoch := strings.Index(str, ":")
	if sepepoch > -1 {
		intepoch, err := strconv.Atoi(str[:sepepoch])
		if err == nil {
			version.epoch = intepoch
		} else {
			return Version{}, errors.New("epoch in version is not a number")
		}
		if intepoch < 0 {
			return Version{}, errors.New("epoch in version is negative")
		}
	} else {
		version.epoch = 0
	}

	// Find version / revision
	seprevision := strings.LastIndex(str, "-")
	if seprevision > -1 {
		version.version = str[sepepoch+1 : seprevision]
		version.revision = str[seprevision+1:]
	} else {
		version.version = str[sepepoch+1:]
		version.revision = ""
	}
	// Verify format
	if len(version.version) == 0 {
		return Version{}, errors.New("No version")
	}

	if !unicode.IsDigit(rune(version.version[0])) {
		return Version{}, errors.New("version does not start with digit")
	}

	for i := 0; i < len(version.version); i = i + 1 {
		r := rune(version.version[i])
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) && !containsRune(versionAllowedSymbols, r) {
			return Version{}, errors.New("invalid character in version")
		}
	}

	for i := 0; i < len(version.revision); i = i + 1 {
		r := rune(version.revision[i])
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) && !containsRune(revisionAllowedSymbols, r) {
			return Version{}, errors.New("invalid character in revision")
		}
	}

	return version, nil
}

// NewVersionUnsafe is just a wrapper around NewVersion that ignore potentiel
// parsing error. Useful for test purposes
func NewVersionUnsafe(str string) Version {
	v, _ := NewVersion(str)
	return v
}

// Compare function compares two Debian-like package version
//
// The implementation is based on http://man.he.net/man5/deb-version
// on https://www.debian.org/doc/debian-policy/ch-controlfields.html#s-f-Version
//
// It uses the dpkg-1.17.25's algorithm  (lib/version.c)
func (a Version) Compare(b Version) int {
	// Quick check
	if a == b {
		return 0
	}

	// Max/Min comparison
	if a == MinVersion || b == MaxVersion {
		return -1
	}
	if b == MinVersion || a == MaxVersion {
		return 1
	}

	// Compare epochs
	if a.epoch > b.epoch {
		return 1
	}
	if a.epoch < b.epoch {
		return -1
	}

	// Compare version
	rc := verrevcmp(a.version, b.version)
	if rc != 0 {
		return signum(rc)
	}

	// Compare revision
	return signum(verrevcmp(a.revision, b.revision))
}

// String returns the string representation of a Version
func (v Version) String() (s string) {
	if v.epoch != 0 {
		s = strconv.Itoa(v.epoch) + ":"
	}
	s += v.version
	if v.revision != "" {
		s += "-" + v.revision
	}
	return
}

func (v Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v *Version) UnmarshalJSON(b []byte) (err error) {
	var str string
	json.Unmarshal(b, &str)
	vp := NewVersionUnsafe(str)
	*v = vp
	return
}

func (v *Version) Scan(value interface{}) (err error) {
	val, ok := value.([]byte)
	if !ok {
		return errors.New("could not scan a Version from a non-string input")
	}
	*v, err = NewVersion(string(val))
	return
}

func (v *Version) Value() (driver.Value, error) {
	return v.String(), nil
}

func verrevcmp(t1, t2 string) int {
	t1, rt1 := nextRune(t1)
	t2, rt2 := nextRune(t2)

	for rt1 != nil || rt2 != nil {
		firstDiff := 0

		for (rt1 != nil && !unicode.IsDigit(*rt1)) || (rt2 != nil && !unicode.IsDigit(*rt2)) {
			ac := 0
			bc := 0
			if rt1 != nil {
				ac = order(*rt1)
			}
			if rt2 != nil {
				bc = order(*rt2)
			}

			if ac != bc {
				return ac - bc
			}

			t1, rt1 = nextRune(t1)
			t2, rt2 = nextRune(t2)
		}
		for rt1 != nil && *rt1 == '0' {
			t1, rt1 = nextRune(t1)
		}
		for rt2 != nil && *rt2 == '0' {
			t2, rt2 = nextRune(t2)
		}
		for rt1 != nil && unicode.IsDigit(*rt1) && rt2 != nil && unicode.IsDigit(*rt2) {
			if firstDiff == 0 {
				firstDiff = int(*rt1) - int(*rt2)
			}
			t1, rt1 = nextRune(t1)
			t2, rt2 = nextRune(t2)
		}
		if rt1 != nil && unicode.IsDigit(*rt1) {
			return 1
		}
		if rt2 != nil && unicode.IsDigit(*rt2) {
			return -1
		}
		if firstDiff != 0 {
			return firstDiff
		}
	}

	return 0
}

// order compares runes using a modified ASCII table
// so that letters are sorted earlier than non-letters
// and so that tildes sorts before anything
func order(r rune) int {
	if unicode.IsDigit(r) {
		return 0
	}

	if unicode.IsLetter(r) {
		return int(r)
	}

	if r == '~' {
		return -1
	}

	return int(r) + 256
}

func nextRune(str string) (string, *rune) {
	if len(str) >= 1 {
		r := rune(str[0])
		return str[1:], &r
	}
	return str, nil
}

func containsRune(s []rune, e rune) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func signum(a int) int {
	switch {
	case a < 0:
		return -1
	case a > 0:
		return +1
	}

	return 0
}
