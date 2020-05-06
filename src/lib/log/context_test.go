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

package log

import (
	"context"
	"testing"
)

func TestGetLogger(t *testing.T) {
	var (
		expectedLevel = ErrorLevel.string()
		expectLine    = "context_test.go:32"
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	G(context.TODO()).Errorf("%s", message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestGetLoggerWithFields(t *testing.T) {
	var (
		expectedLevel = ErrorLevel.string()
		expectLine    = "context_test.go:50"
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	G(context.TODO()).WithFields(Fields{"action": "test"}).Errorf("%s", message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}

	if !contains(t, str, expectedLevel, expectLine, `action="test"`) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestWithLogger(t *testing.T) {
	var (
		expectedLevel = ErrorLevel.string()
		expectLine    = "context_test.go:74"
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	ctx := WithLogger(context.TODO(), L.WithFields(Fields{"action": "test"}))

	G(ctx).WithFields(Fields{"action": "test"}).Errorf("%s", message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}

	if !contains(t, str, expectedLevel, expectLine, `action="test"`) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}
