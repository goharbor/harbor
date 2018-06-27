package migrate

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"testing"

	dStub "github.com/golang-migrate/migrate/database/stub"
	"github.com/golang-migrate/migrate/source"
	sStub "github.com/golang-migrate/migrate/source/stub"
)

// sourceStubMigrations hold the following migrations:
// u = up migration, d = down migration, n = version
//  |  1  |  -  |  3  |  4  |  5  |  -  |  7  |
//  | u d |  -  | u   | u d |   d |  -  | u d |
var sourceStubMigrations *source.Migrations

func init() {
	sourceStubMigrations = source.NewMigrations()
	sourceStubMigrations.Append(&source.Migration{Version: 1, Direction: source.Up})
	sourceStubMigrations.Append(&source.Migration{Version: 1, Direction: source.Down})
	sourceStubMigrations.Append(&source.Migration{Version: 3, Direction: source.Up})
	sourceStubMigrations.Append(&source.Migration{Version: 4, Direction: source.Up})
	sourceStubMigrations.Append(&source.Migration{Version: 4, Direction: source.Down})
	sourceStubMigrations.Append(&source.Migration{Version: 5, Direction: source.Down})
	sourceStubMigrations.Append(&source.Migration{Version: 7, Direction: source.Up})
	sourceStubMigrations.Append(&source.Migration{Version: 7, Direction: source.Down})
}

type DummyInstance struct{ Name string }

func TestNew(t *testing.T) {
	m, err := New("stub://", "stub://")
	if err != nil {
		t.Fatal(err)
	}

	if m.sourceName != "stub" {
		t.Errorf("expected stub, got %v", m.sourceName)
	}
	if m.sourceDrv == nil {
		t.Error("expected sourceDrv not to be nil")
	}

	if m.databaseName != "stub" {
		t.Errorf("expected stub, got %v", m.databaseName)
	}
	if m.databaseDrv == nil {
		t.Error("expected databaseDrv not to be nil")
	}
}

func ExampleNew() {
	// Read migrations from /home/mattes/migrations and connect to a local postgres database.
	m, err := New("file:///home/mattes/migrations", "postgres://mattes:secret@localhost:5432/database?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// Migrate all the way up ...
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
}

func TestNewWithDatabaseInstance(t *testing.T) {
	dummyDb := &DummyInstance{"database"}
	dbInst, err := dStub.WithInstance(dummyDb, &dStub.Config{})
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewWithDatabaseInstance("stub://", "stub", dbInst)
	if err != nil {
		t.Fatal(err)
	}

	if m.sourceName != "stub" {
		t.Errorf("expected stub, got %v", m.sourceName)
	}
	if m.sourceDrv == nil {
		t.Error("expected sourceDrv not to be nil")
	}

	if m.databaseName != "stub" {
		t.Errorf("expected stub, got %v", m.databaseName)
	}
	if m.databaseDrv == nil {
		t.Error("expected databaseDrv not to be nil")
	}
}

func ExampleNewWithDatabaseInstance() {
	// Create and use an existing database instance.
	db, err := sql.Open("postgres", "postgres://mattes:secret@localhost:5432/database?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create driver instance from db.
	// Check each driver if it supports the WithInstance function.
	// `import "github.com/golang-migrate/migrate/database/postgres"`
	instance, err := dStub.WithInstance(db, &dStub.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Read migrations from /home/mattes/migrations and connect to a local postgres database.
	m, err := NewWithDatabaseInstance("file:///home/mattes/migrations", "postgres", instance)
	if err != nil {
		log.Fatal(err)
	}

	// Migrate all the way up ...
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
}

func TestNewWithSourceInstance(t *testing.T) {
	dummySource := &DummyInstance{"source"}
	sInst, err := sStub.WithInstance(dummySource, &sStub.Config{})
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewWithSourceInstance("stub", sInst, "stub://")
	if err != nil {
		t.Fatal(err)
	}

	if m.sourceName != "stub" {
		t.Errorf("expected stub, got %v", m.sourceName)
	}
	if m.sourceDrv == nil {
		t.Error("expected sourceDrv not to be nil")
	}

	if m.databaseName != "stub" {
		t.Errorf("expected stub, got %v", m.databaseName)
	}
	if m.databaseDrv == nil {
		t.Error("expected databaseDrv not to be nil")
	}
}

func ExampleNewWithSourceInstance() {
	di := &DummyInstance{"think any client required for a source here"}

	// Create driver instance from DummyInstance di.
	// Check each driver if it support the WithInstance function.
	// `import "github.com/golang-migrate/migrate/source/stub"`
	instance, err := sStub.WithInstance(di, &sStub.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Read migrations from Stub and connect to a local postgres database.
	m, err := NewWithSourceInstance("stub", instance, "postgres://mattes:secret@localhost:5432/database?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// Migrate all the way up ...
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
}

func TestNewWithInstance(t *testing.T) {
	dummyDb := &DummyInstance{"database"}
	dbInst, err := dStub.WithInstance(dummyDb, &dStub.Config{})
	if err != nil {
		t.Fatal(err)
	}

	dummySource := &DummyInstance{"source"}
	sInst, err := sStub.WithInstance(dummySource, &sStub.Config{})
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewWithInstance("stub", sInst, "stub", dbInst)
	if err != nil {
		t.Fatal(err)
	}

	if m.sourceName != "stub" {
		t.Errorf("expected stub, got %v", m.sourceName)
	}
	if m.sourceDrv == nil {
		t.Error("expected sourceDrv not to be nil")
	}

	if m.databaseName != "stub" {
		t.Errorf("expected stub, got %v", m.databaseName)
	}
	if m.databaseDrv == nil {
		t.Error("expected databaseDrv not to be nil")
	}
}

func ExampleNewWithInstance() {
	// See NewWithDatabaseInstance and NewWithSourceInstance for an example.
}

func TestClose(t *testing.T) {
	m, _ := New("stub://", "stub://")
	sourceErr, databaseErr := m.Close()
	if sourceErr != nil {
		t.Error(sourceErr)
	}
	if databaseErr != nil {
		t.Error(databaseErr)
	}
}

func TestMigrate(t *testing.T) {
	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations
	dbDrv := m.databaseDrv.(*dStub.Stub)
	seq := newMigSeq()

	tt := []struct {
		version       uint
		expectErr     error
		expectVersion uint
		expectSeq     migrationSequence
	}{
		// migrate all the way Up in single steps
		{version: 0, expectErr: os.ErrNotExist},
		{version: 1, expectErr: nil, expectVersion: 1, expectSeq: seq.add(M(1))},
		{version: 2, expectErr: os.ErrNotExist},
		{version: 3, expectErr: nil, expectVersion: 3, expectSeq: seq.add(M(3))},
		{version: 4, expectErr: nil, expectVersion: 4, expectSeq: seq.add(M(4))},
		{version: 5, expectErr: nil, expectVersion: 5, expectSeq: seq.add()}, // 5 has no up migration
		{version: 6, expectErr: os.ErrNotExist},
		{version: 7, expectErr: nil, expectVersion: 7, expectSeq: seq.add(M(7))},
		{version: 8, expectErr: os.ErrNotExist},

		// migrate all the way Down in single steps
		{version: 6, expectErr: os.ErrNotExist},
		{version: 5, expectErr: nil, expectVersion: 5, expectSeq: seq.add(M(7, 5))},
		{version: 4, expectErr: nil, expectVersion: 4, expectSeq: seq.add(M(5, 4))},
		{version: 3, expectErr: nil, expectVersion: 3, expectSeq: seq.add(M(4, 3))},
		{version: 2, expectErr: os.ErrNotExist},
		{version: 1, expectErr: nil, expectVersion: 1, expectSeq: seq.add()}, // 3 has no down migration
		{version: 0, expectErr: os.ErrNotExist},

		// migrate all the way Up in one step
		{version: 7, expectErr: nil, expectVersion: 7, expectSeq: seq.add(M(3), M(4), M(7))},

		// migrate all the way Down in one step
		{version: 1, expectErr: nil, expectVersion: 1, expectSeq: seq.add(M(7, 5), M(5, 4), M(4, 3), M(3, 1))},

		// can't migrate the same version twice
		{version: 1, expectErr: ErrNoChange},
	}

	for i, v := range tt {
		err := m.Migrate(v.version)
		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && err != v.expectErr) {
			t.Errorf("expected err %v, got %v, in %v", v.expectErr, err, i)

		} else if err == nil {
			version, _, err := m.Version()
			if err != nil {
				t.Error(err)
			}
			if version != v.expectVersion {
				t.Errorf("expected version %v, got %v, in %v", v.expectVersion, version, i)
			}
			equalDbSeq(t, i, v.expectSeq, dbDrv)
		}
	}
}

func TestMigrateDirty(t *testing.T) {
	m, _ := New("stub://", "stub://")
	dbDrv := m.databaseDrv.(*dStub.Stub)
	if err := dbDrv.SetVersion(0, true); err != nil {
		t.Fatal(err)
	}

	err := m.Migrate(1)
	if _, ok := err.(ErrDirty); !ok {
		t.Fatalf("expected ErrDirty, got %v", err)
	}
}

func TestSteps(t *testing.T) {
	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations
	dbDrv := m.databaseDrv.(*dStub.Stub)
	seq := newMigSeq()

	tt := []struct {
		n             int
		expectErr     error
		expectVersion int
		expectSeq     migrationSequence
	}{
		// step must be != 0
		{n: 0, expectErr: ErrNoChange},

		// can't go Down if ErrNilVersion
		{n: -1, expectErr: os.ErrNotExist},

		// migrate all the way Up
		{n: 1, expectErr: nil, expectVersion: 1, expectSeq: seq.add(M(1))},
		{n: 1, expectErr: nil, expectVersion: 3, expectSeq: seq.add(M(3))},
		{n: 1, expectErr: nil, expectVersion: 4, expectSeq: seq.add(M(4))},
		{n: 1, expectErr: nil, expectVersion: 5, expectSeq: seq.add()},
		{n: 1, expectErr: nil, expectVersion: 7, expectSeq: seq.add(M(7))},
		{n: 1, expectErr: os.ErrNotExist},

		// migrate all the way Down
		{n: -1, expectErr: nil, expectVersion: 5, expectSeq: seq.add(M(7, 5))},
		{n: -1, expectErr: nil, expectVersion: 4, expectSeq: seq.add(M(5, 4))},
		{n: -1, expectErr: nil, expectVersion: 3, expectSeq: seq.add(M(4, 3))},
		{n: -1, expectErr: nil, expectVersion: 1, expectSeq: seq.add(M(3, 1))},
		{n: -1, expectErr: nil, expectVersion: -1, expectSeq: seq.add(M(1, -1))},

		// migrate Up in bigger step
		{n: 4, expectErr: nil, expectVersion: 5, expectSeq: seq.add(M(1), M(3), M(4), M(5))},

		// apply one migration, then reaches out of boundary
		{n: 2, expectErr: ErrShortLimit{1}, expectVersion: 7, expectSeq: seq.add(M(7))},

		// migrate Down in bigger step
		{n: -4, expectErr: nil, expectVersion: 1, expectSeq: seq.add(M(7, 5), M(5, 4), M(4, 3), M(3, 1))},

		// apply one migration, then reaches out of boundary
		{n: -2, expectErr: ErrShortLimit{1}, expectVersion: -1, expectSeq: seq.add(M(1, -1))},
	}

	for i, v := range tt {
		err := m.Steps(v.n)
		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && err != v.expectErr) {
			t.Errorf("expected err %v, got %v, in %v", v.expectErr, err, i)

		} else if err == nil {
			version, _, err := m.Version()
			if err != ErrNilVersion && err != nil {
				t.Error(err)
			}
			if v.expectVersion == -1 && err != ErrNilVersion {
				t.Errorf("expected ErrNilVersion, got %v, in %v", version, i)

			} else if v.expectVersion >= 0 && version != uint(v.expectVersion) {
				t.Errorf("expected version %v, got %v, in %v", v.expectVersion, version, i)
			}
			equalDbSeq(t, i, v.expectSeq, dbDrv)
		}
	}
}

func TestStepsDirty(t *testing.T) {
	m, _ := New("stub://", "stub://")
	dbDrv := m.databaseDrv.(*dStub.Stub)
	if err := dbDrv.SetVersion(0, true); err != nil {
		t.Fatal(err)
	}

	err := m.Steps(1)
	if _, ok := err.(ErrDirty); !ok {
		t.Fatalf("expected ErrDirty, got %v", err)
	}
}

func TestUpAndDown(t *testing.T) {
	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations
	dbDrv := m.databaseDrv.(*dStub.Stub)
	seq := newMigSeq()

	// go Up first
	if err := m.Up(); err != nil {
		t.Fatal(err)
	}
	equalDbSeq(t, 0, seq.add(M(1), M(3), M(4), M(5), M(7)), dbDrv)

	// go Down
	if err := m.Down(); err != nil {
		t.Fatal(err)
	}
	equalDbSeq(t, 1, seq.add(M(7, 5), M(5, 4), M(4, 3), M(3, 1), M(1, -1)), dbDrv)

	// go 1 Up and then all the way Up
	if err := m.Steps(1); err != nil {
		t.Fatal(err)
	}
	if err := m.Up(); err != nil {
		t.Fatal(err)
	}
	equalDbSeq(t, 2, seq.add(M(1), M(3), M(4), M(5), M(7)), dbDrv)

	// go 1 Down and then all the way Down
	if err := m.Steps(-1); err != nil {
		t.Fatal(err)
	}
	if err := m.Down(); err != nil {
		t.Fatal(err)
	}
	equalDbSeq(t, 0, seq.add(M(7, 5), M(5, 4), M(4, 3), M(3, 1), M(1, -1)), dbDrv)
}

func TestUpDirty(t *testing.T) {
	m, _ := New("stub://", "stub://")
	dbDrv := m.databaseDrv.(*dStub.Stub)
	if err := dbDrv.SetVersion(0, true); err != nil {
		t.Fatal(err)
	}

	err := m.Up()
	if _, ok := err.(ErrDirty); !ok {
		t.Fatalf("expected ErrDirty, got %v", err)
	}
}

func TestDownDirty(t *testing.T) {
	m, _ := New("stub://", "stub://")
	dbDrv := m.databaseDrv.(*dStub.Stub)
	if err := dbDrv.SetVersion(0, true); err != nil {
		t.Fatal(err)
	}

	err := m.Down()
	if _, ok := err.(ErrDirty); !ok {
		t.Fatalf("expected ErrDirty, got %v", err)
	}
}

func TestDrop(t *testing.T) {
	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations
	dbDrv := m.databaseDrv.(*dStub.Stub)

	if err := m.Drop(); err != nil {
		t.Fatal(err)
	}

	if dbDrv.MigrationSequence[len(dbDrv.MigrationSequence)-1] != dStub.DROP {
		t.Fatalf("expected database to DROP, got sequence %v", dbDrv.MigrationSequence)
	}
}

func TestVersion(t *testing.T) {
	m, _ := New("stub://", "stub://")
	dbDrv := m.databaseDrv.(*dStub.Stub)

	_, _, err := m.Version()
	if err != ErrNilVersion {
		t.Fatalf("expected ErrNilVersion, got %v", err)
	}

	if err := dbDrv.Run(bytes.NewBufferString("1_up")); err != nil {
		t.Fatal(err)
	}

	if err := dbDrv.SetVersion(1, false); err != nil {
		t.Fatal(err)
	}

	v, _, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}

	if v != 1 {
		t.Fatalf("expected version 1, got %v", v)
	}
}

func TestRun(t *testing.T) {
	m, _ := New("stub://", "stub://")

	mx, err := NewMigration(nil, "", 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Run(mx); err != nil {
		t.Fatal(err)
	}

	v, _, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}

	if v != 2 {
		t.Errorf("expected version 2, got %v", v)
	}
}

func TestRunDirty(t *testing.T) {
	m, _ := New("stub://", "stub://")
	dbDrv := m.databaseDrv.(*dStub.Stub)
	if err := dbDrv.SetVersion(0, true); err != nil {
		t.Fatal(err)
	}

	migr, err := NewMigration(nil, "", 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Run(migr)
	if _, ok := err.(ErrDirty); !ok {
		t.Fatalf("expected ErrDirty, got %v", err)
	}
}

func TestForce(t *testing.T) {
	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations

	if err := m.Force(7); err != nil {
		t.Fatal(err)
	}

	v, dirty, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Errorf("expected dirty to be false")
	}
	if v != 7 {
		t.Errorf("expected version to be 7")
	}
}

func TestForceDirty(t *testing.T) {
	m, _ := New("stub://", "stub://")
	dbDrv := m.databaseDrv.(*dStub.Stub)
	if err := dbDrv.SetVersion(0, true); err != nil {
		t.Fatal(err)
	}

	if err := m.Force(1); err != nil {
		t.Fatal(err)
	}
}

func TestRead(t *testing.T) {
	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations

	tt := []struct {
		from             int
		to               int
		expectErr        error
		expectMigrations migrationSequence
	}{
		{from: -1, to: -1, expectErr: ErrNoChange},
		{from: -1, to: 0, expectErr: os.ErrNotExist},
		{from: -1, to: 1, expectErr: nil, expectMigrations: newMigSeq(M(1))},
		{from: -1, to: 2, expectErr: os.ErrNotExist},
		{from: -1, to: 3, expectErr: nil, expectMigrations: newMigSeq(M(1), M(3))},
		{from: -1, to: 4, expectErr: nil, expectMigrations: newMigSeq(M(1), M(3), M(4))},
		{from: -1, to: 5, expectErr: nil, expectMigrations: newMigSeq(M(1), M(3), M(4), M(5))},
		{from: -1, to: 6, expectErr: os.ErrNotExist},
		{from: -1, to: 7, expectErr: nil, expectMigrations: newMigSeq(M(1), M(3), M(4), M(5), M(7))},
		{from: -1, to: 8, expectErr: os.ErrNotExist},

		{from: 0, to: -1, expectErr: os.ErrNotExist},
		{from: 0, to: 0, expectErr: os.ErrNotExist},
		{from: 0, to: 1, expectErr: os.ErrNotExist},
		{from: 0, to: 2, expectErr: os.ErrNotExist},
		{from: 0, to: 3, expectErr: os.ErrNotExist},
		{from: 0, to: 4, expectErr: os.ErrNotExist},
		{from: 0, to: 5, expectErr: os.ErrNotExist},
		{from: 0, to: 6, expectErr: os.ErrNotExist},
		{from: 0, to: 7, expectErr: os.ErrNotExist},
		{from: 0, to: 8, expectErr: os.ErrNotExist},

		{from: 1, to: -1, expectErr: nil, expectMigrations: newMigSeq(M(1, -1))},
		{from: 1, to: 0, expectErr: os.ErrNotExist},
		{from: 1, to: 1, expectErr: ErrNoChange},
		{from: 1, to: 2, expectErr: os.ErrNotExist},
		{from: 1, to: 3, expectErr: nil, expectMigrations: newMigSeq(M(3))},
		{from: 1, to: 4, expectErr: nil, expectMigrations: newMigSeq(M(3), M(4))},
		{from: 1, to: 5, expectErr: nil, expectMigrations: newMigSeq(M(3), M(4), M(5))},
		{from: 1, to: 6, expectErr: os.ErrNotExist},
		{from: 1, to: 7, expectErr: nil, expectMigrations: newMigSeq(M(3), M(4), M(5), M(7))},
		{from: 1, to: 8, expectErr: os.ErrNotExist},

		{from: 2, to: -1, expectErr: os.ErrNotExist},
		{from: 2, to: 0, expectErr: os.ErrNotExist},
		{from: 2, to: 1, expectErr: os.ErrNotExist},
		{from: 2, to: 2, expectErr: os.ErrNotExist},
		{from: 2, to: 3, expectErr: os.ErrNotExist},
		{from: 2, to: 4, expectErr: os.ErrNotExist},
		{from: 2, to: 5, expectErr: os.ErrNotExist},
		{from: 2, to: 6, expectErr: os.ErrNotExist},
		{from: 2, to: 7, expectErr: os.ErrNotExist},
		{from: 2, to: 8, expectErr: os.ErrNotExist},

		{from: 3, to: -1, expectErr: nil, expectMigrations: newMigSeq(M(3, 1), M(1, -1))},
		{from: 3, to: 0, expectErr: os.ErrNotExist},
		{from: 3, to: 1, expectErr: nil, expectMigrations: newMigSeq(M(3, 1))},
		{from: 3, to: 2, expectErr: os.ErrNotExist},
		{from: 3, to: 3, expectErr: ErrNoChange},
		{from: 3, to: 4, expectErr: nil, expectMigrations: newMigSeq(M(4))},
		{from: 3, to: 5, expectErr: nil, expectMigrations: newMigSeq(M(4), M(5))},
		{from: 3, to: 6, expectErr: os.ErrNotExist},
		{from: 3, to: 7, expectErr: nil, expectMigrations: newMigSeq(M(4), M(5), M(7))},
		{from: 3, to: 8, expectErr: os.ErrNotExist},

		{from: 4, to: -1, expectErr: nil, expectMigrations: newMigSeq(M(4, 3), M(3, 1), M(1, -1))},
		{from: 4, to: 0, expectErr: os.ErrNotExist},
		{from: 4, to: 1, expectErr: nil, expectMigrations: newMigSeq(M(4, 3), M(3, 1))},
		{from: 4, to: 2, expectErr: os.ErrNotExist},
		{from: 4, to: 3, expectErr: nil, expectMigrations: newMigSeq(M(4, 3))},
		{from: 4, to: 4, expectErr: ErrNoChange},
		{from: 4, to: 5, expectErr: nil, expectMigrations: newMigSeq(M(5))},
		{from: 4, to: 6, expectErr: os.ErrNotExist},
		{from: 4, to: 7, expectErr: nil, expectMigrations: newMigSeq(M(5), M(7))},
		{from: 4, to: 8, expectErr: os.ErrNotExist},

		{from: 5, to: -1, expectErr: nil, expectMigrations: newMigSeq(M(5, 4), M(4, 3), M(3, 1), M(1, -1))},
		{from: 5, to: 0, expectErr: os.ErrNotExist},
		{from: 5, to: 1, expectErr: nil, expectMigrations: newMigSeq(M(5, 4), M(4, 3), M(3, 1))},
		{from: 5, to: 2, expectErr: os.ErrNotExist},
		{from: 5, to: 3, expectErr: nil, expectMigrations: newMigSeq(M(5, 4), M(4, 3))},
		{from: 5, to: 4, expectErr: nil, expectMigrations: newMigSeq(M(5, 4))},
		{from: 5, to: 5, expectErr: ErrNoChange},
		{from: 5, to: 6, expectErr: os.ErrNotExist},
		{from: 5, to: 7, expectErr: nil, expectMigrations: newMigSeq(M(7))},
		{from: 5, to: 8, expectErr: os.ErrNotExist},

		{from: 6, to: -1, expectErr: os.ErrNotExist},
		{from: 6, to: 0, expectErr: os.ErrNotExist},
		{from: 6, to: 1, expectErr: os.ErrNotExist},
		{from: 6, to: 2, expectErr: os.ErrNotExist},
		{from: 6, to: 3, expectErr: os.ErrNotExist},
		{from: 6, to: 4, expectErr: os.ErrNotExist},
		{from: 6, to: 5, expectErr: os.ErrNotExist},
		{from: 6, to: 6, expectErr: os.ErrNotExist},
		{from: 6, to: 7, expectErr: os.ErrNotExist},
		{from: 6, to: 8, expectErr: os.ErrNotExist},

		{from: 7, to: -1, expectErr: nil, expectMigrations: newMigSeq(M(7, 5), M(5, 4), M(4, 3), M(3, 1), M(1, -1))},
		{from: 7, to: 0, expectErr: os.ErrNotExist},
		{from: 7, to: 1, expectErr: nil, expectMigrations: newMigSeq(M(7, 5), M(5, 4), M(4, 3), M(3, 1))},
		{from: 7, to: 2, expectErr: os.ErrNotExist},
		{from: 7, to: 3, expectErr: nil, expectMigrations: newMigSeq(M(7, 5), M(5, 4), M(4, 3))},
		{from: 7, to: 4, expectErr: nil, expectMigrations: newMigSeq(M(7, 5), M(5, 4))},
		{from: 7, to: 5, expectErr: nil, expectMigrations: newMigSeq(M(7, 5))},
		{from: 7, to: 6, expectErr: os.ErrNotExist},
		{from: 7, to: 7, expectErr: ErrNoChange},
		{from: 7, to: 8, expectErr: os.ErrNotExist},

		{from: 8, to: -1, expectErr: os.ErrNotExist},
		{from: 8, to: 0, expectErr: os.ErrNotExist},
		{from: 8, to: 1, expectErr: os.ErrNotExist},
		{from: 8, to: 2, expectErr: os.ErrNotExist},
		{from: 8, to: 3, expectErr: os.ErrNotExist},
		{from: 8, to: 4, expectErr: os.ErrNotExist},
		{from: 8, to: 5, expectErr: os.ErrNotExist},
		{from: 8, to: 6, expectErr: os.ErrNotExist},
		{from: 8, to: 7, expectErr: os.ErrNotExist},
		{from: 8, to: 8, expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		ret := make(chan interface{})
		go m.read(v.from, v.to, ret)
		migrations, err := migrationsFromChannel(ret)

		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && v.expectErr != err) {
			t.Errorf("expected %v, got %v, in %v", v.expectErr, err, i)
			t.Logf("%v, in %v", migrations, i)
		}
		if len(v.expectMigrations) > 0 {
			equalMigSeq(t, i, v.expectMigrations, migrations)
		}
	}
}

func TestReadUp(t *testing.T) {
	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations

	tt := []struct {
		from             int
		limit            int // -1 means no limit
		expectErr        error
		expectMigrations migrationSequence
	}{
		{from: -1, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(1), M(3), M(4), M(5), M(7))},
		{from: -1, limit: 0, expectErr: ErrNoChange},
		{from: -1, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(1))},
		{from: -1, limit: 2, expectErr: nil, expectMigrations: newMigSeq(M(1), M(3))},

		{from: 0, limit: -1, expectErr: os.ErrNotExist},
		{from: 0, limit: 0, expectErr: os.ErrNotExist},
		{from: 0, limit: 1, expectErr: os.ErrNotExist},
		{from: 0, limit: 2, expectErr: os.ErrNotExist},

		{from: 1, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(3), M(4), M(5), M(7))},
		{from: 1, limit: 0, expectErr: ErrNoChange},
		{from: 1, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(3))},
		{from: 1, limit: 2, expectErr: nil, expectMigrations: newMigSeq(M(3), M(4))},

		{from: 2, limit: -1, expectErr: os.ErrNotExist},
		{from: 2, limit: 0, expectErr: os.ErrNotExist},
		{from: 2, limit: 1, expectErr: os.ErrNotExist},
		{from: 2, limit: 2, expectErr: os.ErrNotExist},

		{from: 3, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(4), M(5), M(7))},
		{from: 3, limit: 0, expectErr: ErrNoChange},
		{from: 3, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(4))},
		{from: 3, limit: 2, expectErr: nil, expectMigrations: newMigSeq(M(4), M(5))},

		{from: 4, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(5), M(7))},
		{from: 4, limit: 0, expectErr: ErrNoChange},
		{from: 4, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(5))},
		{from: 4, limit: 2, expectErr: nil, expectMigrations: newMigSeq(M(5), M(7))},

		{from: 5, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(7))},
		{from: 5, limit: 0, expectErr: ErrNoChange},
		{from: 5, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(7))},
		{from: 5, limit: 2, expectErr: ErrShortLimit{1}, expectMigrations: newMigSeq(M(7))},

		{from: 6, limit: -1, expectErr: os.ErrNotExist},
		{from: 6, limit: 0, expectErr: os.ErrNotExist},
		{from: 6, limit: 1, expectErr: os.ErrNotExist},
		{from: 6, limit: 2, expectErr: os.ErrNotExist},

		{from: 7, limit: -1, expectErr: ErrNoChange},
		{from: 7, limit: 0, expectErr: ErrNoChange},
		{from: 7, limit: 1, expectErr: os.ErrNotExist},
		{from: 7, limit: 2, expectErr: os.ErrNotExist},

		{from: 8, limit: -1, expectErr: os.ErrNotExist},
		{from: 8, limit: 0, expectErr: os.ErrNotExist},
		{from: 8, limit: 1, expectErr: os.ErrNotExist},
		{from: 8, limit: 2, expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		ret := make(chan interface{})
		go m.readUp(v.from, v.limit, ret)
		migrations, err := migrationsFromChannel(ret)

		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && v.expectErr != err) {
			t.Errorf("expected %v, got %v, in %v", v.expectErr, err, i)
			t.Logf("%v, in %v", migrations, i)
		}
		if len(v.expectMigrations) > 0 {
			equalMigSeq(t, i, v.expectMigrations, migrations)
		}
	}
}

func TestReadDown(t *testing.T) {
	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations

	tt := []struct {
		from             int
		limit            int // -1 means no limit
		expectErr        error
		expectMigrations migrationSequence
	}{
		{from: -1, limit: -1, expectErr: ErrNoChange},
		{from: -1, limit: 0, expectErr: ErrNoChange},
		{from: -1, limit: 1, expectErr: os.ErrNotExist},
		{from: -1, limit: 2, expectErr: os.ErrNotExist},

		{from: 0, limit: -1, expectErr: os.ErrNotExist},
		{from: 0, limit: 0, expectErr: os.ErrNotExist},
		{from: 0, limit: 1, expectErr: os.ErrNotExist},
		{from: 0, limit: 2, expectErr: os.ErrNotExist},

		{from: 1, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(1, -1))},
		{from: 1, limit: 0, expectErr: ErrNoChange},
		{from: 1, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(1, -1))},
		{from: 1, limit: 2, expectErr: ErrShortLimit{1}, expectMigrations: newMigSeq(M(1, -1))},

		{from: 2, limit: -1, expectErr: os.ErrNotExist},
		{from: 2, limit: 0, expectErr: os.ErrNotExist},
		{from: 2, limit: 1, expectErr: os.ErrNotExist},
		{from: 2, limit: 2, expectErr: os.ErrNotExist},

		{from: 3, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(3, 1), M(1, -1))},
		{from: 3, limit: 0, expectErr: ErrNoChange},
		{from: 3, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(3, 1))},
		{from: 3, limit: 2, expectErr: nil, expectMigrations: newMigSeq(M(3, 1), M(1, -1))},

		{from: 4, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(4, 3), M(3, 1), M(1, -1))},
		{from: 4, limit: 0, expectErr: ErrNoChange},
		{from: 4, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(4, 3))},
		{from: 4, limit: 2, expectErr: nil, expectMigrations: newMigSeq(M(4, 3), M(3, 1))},

		{from: 5, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(5, 4), M(4, 3), M(3, 1), M(1, -1))},
		{from: 5, limit: 0, expectErr: ErrNoChange},
		{from: 5, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(5, 4))},
		{from: 5, limit: 2, expectErr: nil, expectMigrations: newMigSeq(M(5, 4), M(4, 3))},

		{from: 6, limit: -1, expectErr: os.ErrNotExist},
		{from: 6, limit: 0, expectErr: os.ErrNotExist},
		{from: 6, limit: 1, expectErr: os.ErrNotExist},
		{from: 6, limit: 2, expectErr: os.ErrNotExist},

		{from: 7, limit: -1, expectErr: nil, expectMigrations: newMigSeq(M(7, 5), M(5, 4), M(4, 3), M(3, 1), M(1, -1))},
		{from: 7, limit: 0, expectErr: ErrNoChange},
		{from: 7, limit: 1, expectErr: nil, expectMigrations: newMigSeq(M(7, 5))},
		{from: 7, limit: 2, expectErr: nil, expectMigrations: newMigSeq(M(7, 5), M(5, 4))},

		{from: 8, limit: -1, expectErr: os.ErrNotExist},
		{from: 8, limit: 0, expectErr: os.ErrNotExist},
		{from: 8, limit: 1, expectErr: os.ErrNotExist},
		{from: 8, limit: 2, expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		ret := make(chan interface{})
		go m.readDown(v.from, v.limit, ret)
		migrations, err := migrationsFromChannel(ret)

		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && v.expectErr != err) {
			t.Errorf("expected %v, got %v, in %v", v.expectErr, err, i)
			t.Logf("%v, in %v", migrations, i)
		}
		if len(v.expectMigrations) > 0 {
			equalMigSeq(t, i, v.expectMigrations, migrations)
		}
	}
}

func TestLock(t *testing.T) {
	m, _ := New("stub://", "stub://")
	if err := m.lock(); err != nil {
		t.Fatal(err)
	}

	if err := m.lock(); err == nil {
		t.Fatal("should be locked already")
	}
}

func migrationsFromChannel(ret chan interface{}) ([]*Migration, error) {
	slice := make([]*Migration, 0)
	for r := range ret {
		switch r.(type) {
		case error:
			return slice, r.(error)

		case *Migration:
			slice = append(slice, r.(*Migration))
		}
	}
	return slice, nil
}

type migrationSequence []*Migration

func newMigSeq(migr ...*Migration) migrationSequence {
	return migr
}

func (m *migrationSequence) add(migr ...*Migration) migrationSequence {
	*m = append(*m, migr...)
	return *m
}

func (m *migrationSequence) bodySequence() []string {
	r := make([]string, 0)
	for _, v := range *m {
		if v.Body != nil {
			body, err := ioutil.ReadAll(v.Body)
			if err != nil {
				panic(err) // that should never happen
			}

			// reset body reader
			// TODO: is there a better/nicer way?
			v.Body = ioutil.NopCloser(bytes.NewReader(body))

			r = append(r, string(body[:]))
		}
	}
	return r
}

// M is a convenience func to create a new *Migration
func M(version uint, targetVersion ...int) *Migration {
	if len(targetVersion) > 1 {
		panic("only one targetVersion allowed")
	}
	ts := int(version)
	if len(targetVersion) == 1 {
		ts = targetVersion[0]
	}

	m, _ := New("stub://", "stub://")
	m.sourceDrv.(*sStub.Stub).Migrations = sourceStubMigrations
	migr, err := m.newMigration(version, ts)
	if err != nil {
		panic(err)
	}
	return migr
}

func equalMigSeq(t *testing.T, i int, expected, got migrationSequence) {
	if len(expected) != len(got) {
		t.Errorf("expected migrations %v, got %v, in %v", expected, got, i)

	} else {
		for ii := 0; ii < len(expected); ii++ {
			if expected[ii].Version != got[ii].Version {
				t.Errorf("expected version %v, got %v, in %v", expected[ii].Version, got[ii].Version, i)
			}

			if expected[ii].TargetVersion != got[ii].TargetVersion {
				t.Errorf("expected targetVersion %v, got %v, in %v", expected[ii].TargetVersion, got[ii].TargetVersion, i)
			}
		}
	}
}

func equalDbSeq(t *testing.T, i int, expected migrationSequence, got *dStub.Stub) {
	bs := expected.bodySequence()
	if !got.EqualSequence(bs) {
		t.Fatalf("\nexpected sequence %v,\ngot               %v, in %v", bs, got.MigrationSequence, i)
	}
}
