// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"strings"
)

// injection is a helper type to manage injected SQLs in all builders.
type injection struct {
	markerSQLs map[injectionMarker][]string
}

type injectionMarker int

type injectedSQL struct {
	marker injectionMarker
	sql    string
}

// newInjection creates a new injection.
func newInjection() *injection {
	return &injection{
		markerSQLs: map[injectionMarker][]string{},
	}
}

// SQL adds sql to injection's sql list.
// All sqls inside injection is ordered by marker in ascending order.
func (injection *injection) SQL(marker injectionMarker, sql string) {
	injection.markerSQLs[marker] = append(injection.markerSQLs[marker], sql)
}

// WriteTo joins all SQL strings at the same marker value with blank (" ")
// and writes the joined value to buf.
func (injection *injection) WriteTo(buf *bytes.Buffer, marker injectionMarker) {
	sqls := injection.markerSQLs[marker]
	notEmpty := buf.Len() > 0

	if len(sqls) == 0 {
		return
	}

	if notEmpty {
		buf.WriteRune(' ')
	}

	s := strings.Join(sqls, " ")
	buf.WriteString(s)

	if !notEmpty {
		buf.WriteRune(' ')
	}
}
