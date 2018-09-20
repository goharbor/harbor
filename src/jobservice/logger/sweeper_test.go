// Copyright Project Harbor Authors. All rights reserved.
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
