package stub

import (
	"testing"

	"github.com/golang-migrate/migrate/source"
	st "github.com/golang-migrate/migrate/source/testing"
)

func Test(t *testing.T) {
	s := &Stub{}
	d, err := s.Open("")
	if err != nil {
		t.Fatal(err)
	}

	m := source.NewMigrations()
	m.Append(&source.Migration{Version: 1, Direction: source.Up})
	m.Append(&source.Migration{Version: 1, Direction: source.Down})
	m.Append(&source.Migration{Version: 3, Direction: source.Up})
	m.Append(&source.Migration{Version: 4, Direction: source.Up})
	m.Append(&source.Migration{Version: 4, Direction: source.Down})
	m.Append(&source.Migration{Version: 5, Direction: source.Down})
	m.Append(&source.Migration{Version: 7, Direction: source.Up})
	m.Append(&source.Migration{Version: 7, Direction: source.Down})

	d.(*Stub).Migrations = m

	st.Test(t, d)
}
