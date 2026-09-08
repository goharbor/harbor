package controllers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSessionType(t *testing.T) {
	tests := []struct {
		name          string
		refreshToken  string
		expectedType  string
		expectedError bool
	}{
		{
			name:          "Valid",
			refreshToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXAiOiJvZmZsaW5lIn0.d9fcdba7c10fc1263bf682947afabaecf3496070cd2d5a5e7b3c79dbf1545c1f",
			expectedType:  "offline",
			expectedError: false,
		},
		{
			name:          "Invalid",
			refreshToken:  "invalidToken",
			expectedType:  "",
			expectedError: true,
		},
		{
			name:          "Missing 'typ' claim",
			refreshToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJIUzI1NiJ9.d9fcdba7c10fc1263bf682947afabaecf3496070cd2d5a5e7b3c79dbf1545c1f",
			expectedType:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := getSessionType(tt.refreshToken)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedType, typ)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, typ)
			}
		})
	}
}
