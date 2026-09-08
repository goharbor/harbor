package lib

import (
	"os"
	"testing"
)

func TestGetEnvInt64(t *testing.T) {
	tests := []struct {
		name          string
		envKey        string
		envValue      string
		defaultValue  int64
		setEnv        bool
		expectedValue int64
	}{
		{
			name:          "env set with valid value",
			envKey:        "TEST_ENV",
			envValue:      "100",
			defaultValue:  50,
			setEnv:        true,
			expectedValue: 100,
		},
		{
			name:          "env not set",
			envKey:        "UNSET_ENV",
			envValue:      "",
			defaultValue:  50,
			setEnv:        false,
			expectedValue: 50,
		},
		{
			name:          "env set with invalid value",
			envKey:        "INVALID_ENV",
			envValue:      "not_a_number",
			defaultValue:  50,
			setEnv:        true,
			expectedValue: 50,
		},
		{
			name:          "env set with zero",
			envKey:        "ZERO_ENV",
			envValue:      "0",
			defaultValue:  50,
			setEnv:        true,
			expectedValue: 0,
		},
		{
			name:          "env set with negative value",
			envKey:        "NEGATIVE_ENV",
			envValue:      "-10",
			defaultValue:  50,
			setEnv:        true,
			expectedValue: -10,
		},
		{
			name:          "env set with large value",
			envKey:        "LARGE_ENV",
			envValue:      "9223372036854775807",
			defaultValue:  50,
			setEnv:        true,
			expectedValue: 9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.envKey, tt.envValue)
			} else {
				os.Unsetenv(tt.envKey)
			}

			result := GetEnvInt64(tt.envKey, tt.defaultValue)
			if result != tt.expectedValue {
				t.Errorf("GetEnvInt64(%q, %d) = %d; want %d", tt.envKey, tt.defaultValue, result, tt.expectedValue)
			}

			os.Unsetenv(tt.envKey)
		})
	}
}
