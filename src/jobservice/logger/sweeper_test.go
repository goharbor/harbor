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
package logger

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSweeper(t *testing.T) {
	workDir := "/tmp/sweeper_logs"

	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	_, err := os.Create(fmt.Sprintf("%s/sweeper_test.log", workDir))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sweeper := NewSweeper(ctx, workDir, 1)
	sweeper.Start()
	<-time.After(100 * time.Millisecond)

	if err := os.Remove(fmt.Sprintf("%s/sweeper_test.log", workDir)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(workDir); err != nil {
		t.Fatal(err)
	}
}
