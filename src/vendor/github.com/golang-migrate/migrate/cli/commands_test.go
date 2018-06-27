package main

import (
	"testing"
)

func TestNextSeq(t *testing.T) {
	cases := []struct {
		name           string
		matches        []string
		dir            string
		seqDigits      int
		expected       string
		expectedErrStr string
	}{
		{"Bad digits", []string{}, "migrationDir", 0, "", "Digits must be positive"},
		{"Single digit initialize", []string{}, "migrationDir", 1, "1", ""},
		{"Single digit malformed", []string{"bad"}, "migrationDir", 1, "", "Malformed migration filename: bad"},
		{"Single digit no int", []string{"bad_bad"}, "migrationDir", 1, "", "strconv.Atoi: parsing \"bad\": invalid syntax"},
		{"Single digit negative seq", []string{"-5_test"}, "migrationDir", 1, "", "Next sequence number must be positive"},
		{"Single digit increment", []string{"3_test", "4_test"}, "migrationDir", 1, "5", ""},
		{"Single digit overflow", []string{"9_test"}, "migrationDir", 1, "", "Next sequence number 10 too large. At most 1 digits are allowed"},
		{"Zero-pad initialize", []string{}, "migrationDir", 6, "000001", ""},
		{"Zero-pad malformed", []string{"bad"}, "migrationDir", 6, "", "Malformed migration filename: bad"},
		{"Zero-pad no int", []string{"bad_bad"}, "migrationDir", 6, "", "strconv.Atoi: parsing \"bad\": invalid syntax"},
		{"Zero-pad negative seq", []string{"-000005_test"}, "migrationDir", 6, "", "Next sequence number must be positive"},
		{"Zero-pad increment", []string{"000003_test", "000004_test"}, "migrationDir", 6, "000005", ""},
		{"Zero-pad overflow", []string{"999999_test"}, "migrationDir", 6, "", "Next sequence number 1000000 too large. At most 6 digits are allowed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			nextSeq, err := nextSeq(c.matches, c.dir, c.seqDigits)
			if nextSeq != c.expected {
				t.Error("Incorrect nextSeq: " + nextSeq + " != " + c.expected)
			}
			if err != nil {
				if err.Error() != c.expectedErrStr {
					t.Error("Incorrect error: " + err.Error() + " != " + c.expectedErrStr)
				}
			} else if c.expectedErrStr != "" {
				t.Error("Expected error: " + c.expectedErrStr + " but got nil instead")
			}
		})
	}
}
