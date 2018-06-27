package source

import (
	"testing"
)

func TestNewMigrations(t *testing.T) {
	// TODO
}

func TestAppend(t *testing.T) {
	// TODO
}

func TestBuildIndex(t *testing.T) {
	// TODO
}

func TestFirst(t *testing.T) {
	// TODO
}

func TestPrev(t *testing.T) {
	// TODO
}

func TestUp(t *testing.T) {
	// TODO
}

func TestDown(t *testing.T) {
	// TODO
}

func TestFindPos(t *testing.T) {
	m := Migrations{index: uintSlice{1, 2, 3}}
	if p := m.findPos(0); p != -1 {
		t.Errorf("expected -1, got %v", p)
	}
	if p := m.findPos(1); p != 0 {
		t.Errorf("expected 0, got %v", p)
	}
	if p := m.findPos(3); p != 2 {
		t.Errorf("expected 2, got %v", p)
	}
}
