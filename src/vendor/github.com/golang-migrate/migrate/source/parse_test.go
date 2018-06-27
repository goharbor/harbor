package source

import (
	"testing"
)

func TestParse(t *testing.T) {
	tt := []struct {
		name            string
		expectErr       error
		expectMigration *Migration
	}{
		{
			name:      "1_foobar.up.sql",
			expectErr: nil,
			expectMigration: &Migration{
				Version:    1,
				Identifier: "foobar",
				Direction:  Up,
				Raw:        "1_foobar.up.sql",
			},
		},
		{
			name:      "1_foobar.down.sql",
			expectErr: nil,
			expectMigration: &Migration{
				Version:    1,
				Identifier: "foobar",
				Direction:  Down,
				Raw:        "1_foobar.down.sql",
			},
		},
		{
			name:      "1_f-o_ob+ar.up.sql",
			expectErr: nil,
			expectMigration: &Migration{
				Version:    1,
				Identifier: "f-o_ob+ar",
				Direction:  Up,
				Raw:        "1_f-o_ob+ar.up.sql",
			},
		},
		{
			name:      "1485385885_foobar.up.sql",
			expectErr: nil,
			expectMigration: &Migration{
				Version:    1485385885,
				Identifier: "foobar",
				Direction:  Up,
				Raw:        "1485385885_foobar.up.sql",
			},
		},
		{
			name:      "20170412214116_date_foobar.up.sql",
			expectErr: nil,
			expectMigration: &Migration{
				Version:    20170412214116,
				Identifier: "date_foobar",
				Direction:  Up,
				Raw:        "20170412214116_date_foobar.up.sql",
			},
		},
		{
			name:            "-1_foobar.up.sql",
			expectErr:       ErrParse,
			expectMigration: nil,
		},
		{
			name:            "foobar.up.sql",
			expectErr:       ErrParse,
			expectMigration: nil,
		},
		{
			name:            "1.up.sql",
			expectErr:       ErrParse,
			expectMigration: nil,
		},
		{
			name:            "1_foobar.sql",
			expectErr:       ErrParse,
			expectMigration: nil,
		},
		{
			name:            "1_foobar.up",
			expectErr:       ErrParse,
			expectMigration: nil,
		},
		{
			name:            "1_foobar.down",
			expectErr:       ErrParse,
			expectMigration: nil,
		},
	}

	for i, v := range tt {
		f, err := Parse(v.name)

		if err != v.expectErr {
			t.Errorf("expected %v, got %v, in %v", v.expectErr, err, i)
		}

		if v.expectMigration != nil && *f != *v.expectMigration {
			t.Errorf("expected %+v, got %+v, in %v", *v.expectMigration, *f, i)
		}
	}
}
