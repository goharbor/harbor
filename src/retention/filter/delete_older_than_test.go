package filter

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/retention"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestNewDeleteOlderThan(t *testing.T) {
	tests := []struct {
		Name        string
		Metadata    map[string]interface{}
		ExpectedN   int
		ExpectedErr error
	}{
		{
			Name:      "Valid",
			Metadata:  map[string]interface{}{MetaDataKeyN: 3},
			ExpectedN: 3,
		},
		{
			Name:        "Missing N",
			Metadata:    map[string]interface{}{"_": 3},
			ExpectedErr: ErrMissingMetadata(MetaDataKeyN),
		},
		{
			Name:        "N Is Wrong Type",
			Metadata:    map[string]interface{}{MetaDataKeyN: "3"},
			ExpectedErr: ErrWrongMetadataType(MetaDataKeyN, "int"),
		},
		{
			Name:        "N Is Negative",
			Metadata:    map[string]interface{}{MetaDataKeyN: -1},
			ExpectedErr: ErrInvalidMetadata(MetaDataKeyN, "cannot be negative"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			sut, err := NewDeleteOlderThan(tt.Metadata)

			if tt.ExpectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, tt.ExpectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.ExpectedN > 0 {
				require.NotNil(t, sut)
				assert.Equal(t, tt.ExpectedN, sut.n)
			} else {
				assert.Nil(t, sut)
			}
		})
	}
}

func TestDeleteOlderThan_Process(t *testing.T) {
	sut := deleteOlderThan{n: 10}
	now := time.Now()

	tests := []struct {
		CreatedAt      time.Time
		ExpectedDelete bool
	}{
		{CreatedAt: now.Add(1 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(2 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(3 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(4 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(5 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(6 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(7 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(8 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(9 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(10 * daysAgo), ExpectedDelete: false},
		{CreatedAt: now.Add(10 * daysAgo).Add(-1 * time.Second), ExpectedDelete: true},
		{CreatedAt: now.Add(11 * daysAgo), ExpectedDelete: true},
		{CreatedAt: now.Add(12 * daysAgo), ExpectedDelete: true},
	}

	for _, tt := range tests {
		t.Run(now.Sub(tt.CreatedAt).String(), func(t *testing.T) {
			result, err := sut.Process(&retention.TagRecord{CreatedAt: tt.CreatedAt})

			require.NoError(t, err)

			if tt.ExpectedDelete {
				require.Equal(t, retention.FilterActionDelete, result)
			} else {
				require.Equal(t, retention.FilterActionNoDecision, result)
			}
		})
	}
}
