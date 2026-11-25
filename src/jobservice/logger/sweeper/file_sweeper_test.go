package sweeper

import (
	"os"
	"path"
	"testing"
	"time"
)

// Test file sweeper
func TestFileSweeper(t *testing.T) {
	workDir := path.Join(os.TempDir(), "job_logs")
	if err := os.Mkdir(workDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(workDir); err != nil {
			t.Error(err)
		}
	}()

	logFile := path.Join(workDir, "TestFileSweeper.log")
	if err := os.WriteFile(logFile, []byte("hello"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	oldModTime := time.Unix(time.Now().Unix()-6*24*3600, 0)
	if err := os.Chtimes(logFile, oldModTime, oldModTime); err != nil {
		t.Error(err)
	}

	fs := NewFileSweeper(workDir, 5)
	if fs.Duration() != 5 {
		t.Errorf("expect duration 5 but got %d", fs.Duration())
	}

	count, err := fs.Sweep()
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Errorf("expect count 1 but got %d", count)
	}
}

// TestFileSweeper_Thresholds tests retention boundaries: younger, older, and boundary conditions
func TestFileSweeper_Thresholds(t *testing.T) {
	workDir, err := os.MkdirTemp("", "job_logs_threshold_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	now := time.Now()
	retentionDays := 5

	// 1. Older than threshold (6 days old) - SHOULD be deleted
	olderFile := path.Join(workDir, "older.log")
	if err := os.WriteFile(olderFile, []byte("older"), 0600); err != nil {
		t.Fatal(err)
	}
	olderTime := now.Add(-time.Duration(retentionDays+1) * 24 * time.Hour)
	if err := os.Chtimes(olderFile, olderTime, olderTime); err != nil {
		t.Fatal(err)
	}

	// 2. Younger than threshold (3 days old) - SHOULD be retained
	youngerFile := path.Join(workDir, "younger.log")
	if err := os.WriteFile(youngerFile, []byte("younger"), 0600); err != nil {
		t.Fatal(err)
	}
	youngerTime := now.Add(-time.Duration(retentionDays-2) * 24 * time.Hour)
	if err := os.Chtimes(youngerFile, youngerTime, youngerTime); err != nil {
		t.Fatal(err)
	}

	// 3. Exactly at threshold (5 days old plus a buffer so it's not before Now) - SHOULD be retained
	exactFile := path.Join(workDir, "exact.log")
	if err := os.WriteFile(exactFile, []byte("exact"), 0600); err != nil {
		t.Fatal(err)
	}
	// ModTime + 5 days = now + 1 hour -> not Before(now)
	exactTime := now.Add(-time.Duration(retentionDays)*24*time.Hour + 1*time.Hour)
	if err := os.Chtimes(exactFile, exactTime, exactTime); err != nil {
		t.Fatal(err)
	}

	// 4. Fresh file (0 days old) - SHOULD be retained
	freshFile := path.Join(workDir, "fresh.log")
	if err := os.WriteFile(freshFile, []byte("fresh"), 0600); err != nil {
		t.Fatal(err)
	}

	fs := NewFileSweeper(workDir, retentionDays)
	count, err := fs.Sweep()
	if err != nil {
		t.Fatalf("unexpected error during sweep: %v", err)
	}

	// Only the older file should have been removed
	if count != 1 {
		t.Errorf("expected 1 file cleared, got %d", count)
	}

	if _, err := os.Stat(olderFile); !os.IsNotExist(err) {
		t.Errorf("expected older file to be removed, but it still exists")
	}
	if _, err := os.Stat(youngerFile); err != nil {
		t.Errorf("expected younger file to exist, got err: %v", err)
	}
	if _, err := os.Stat(exactFile); err != nil {
		t.Errorf("expected exact boundary file to exist, got err: %v", err)
	}
	if _, err := os.Stat(freshFile); err != nil {
		t.Errorf("expected fresh file to exist, got err: %v", err)
	}
}

func TestFileSweeper_EmptyDirectory(t *testing.T) {
	workDir, err := os.MkdirTemp("", "job_logs_empty_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	fs := NewFileSweeper(workDir, 1)
	count, err := fs.Sweep()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 files cleared, got %d", count)
	}
}

func TestFileSweeper_NonExistentDirectory(t *testing.T) {
	fs := NewFileSweeper("/path/that/does/not/exist/for/sure/12345", 1)
	count, err := fs.Sweep()
	if err == nil {
		t.Fatalf("expected error for non-existent directory, got nil")
	}
	if count != 0 {
		t.Errorf("expected 0 cleared on error, got %d", count)
	}
}
