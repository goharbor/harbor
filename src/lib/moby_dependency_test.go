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

package lib

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestNoDirectMobyDependency verifies that harbor does not directly consume
// the moby module (github.com/docker/docker or github.com/moby/*).
// Moby is a large Docker daemon codebase; harbor should not depend on it directly.
// Indirect/transitive dependencies via other libraries are acceptable, but
// harbor's own go.mod must not list moby as a required module.
func TestNoDirectMobyDependency(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	// go.mod is one directory above lib/
	goModPath := filepath.Join(filepath.Dir(filename), "..", "go.mod")

	f, err := os.Open(goModPath)
	if err != nil {
		t.Fatalf("failed to open go.mod: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and blank lines
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}
		// Extract the module path (first whitespace-separated field) from the line.
		// In a require block each line is "<module> <version> [// indirect]".
		modulePath := strings.Fields(line)[0]
		if strings.HasPrefix(modulePath, "github.com/moby/") || modulePath == "github.com/docker/docker" {
			t.Errorf("harbor go.mod must not directly require the moby module, but found: %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading go.mod: %v", err)
	}
}
