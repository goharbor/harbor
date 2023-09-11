package lib

import "testing"

func TestValidateQuotaLimit(t *testing.T) {
	testCases := []struct {
		description  string
		storageLimit int64
		hasError     bool
	}{
		{
			description:  "storage limit is -2",
			storageLimit: -2,
			hasError:     true,
		},
		{
			description:  "storage limit is -1",
			storageLimit: -1,
			hasError:     false,
		},
		{
			description:  "storage limit is 0",
			storageLimit: 0,
			hasError:     true,
		},
		{
			description:  "storage limit is 1125899906842624",
			storageLimit: 1125899906842624,
			hasError:     false,
		},
		{
			description:  "storage limit is 1125899906842625",
			storageLimit: 1125899906842625,
			hasError:     true,
		},
	}

	for _, tc := range testCases {
		gotErr := ValidateQuotaLimit(tc.storageLimit)
		if tc.hasError {
			if gotErr == nil {
				t.Errorf("test case: %s, it expects error, while got error is nil", tc.description)
			}
		} else {
			// tc.hasError == false
			if gotErr != nil {
				t.Errorf("test case: %s, it doesn't expect error, while got error is not nil, gotErr=%v", tc.description, gotErr)
			}
		}
	}
}
