package postprocessors

import (
	"context"
	mock "github.com/stretchr/testify/mock"
)

// ScanReportV1ToV2Converter is an auto-generated mock type for converting native Harbor report in JSON
// to relational schema
type ScanReportV1ToV2Converter struct {
	mock.Mock
}

// ToRelationalSchema is a mock implementation of the scan report conversion
func (_c *ScanReportV1ToV2Converter) ToRelationalSchema(ctx context.Context, reportUUID string, registrationUUID string, digest string, reportData string) (string, string, error) {
	return "mockId", reportData, nil
}

// ToRelationalSchema is a mock implementation of the scan report conversion
func (_c *ScanReportV1ToV2Converter) FromRelationalSchema(ctx context.Context, reportUUID string, artifactDigest string, reportData string) (string, error) {
	return "mockId", nil
}
