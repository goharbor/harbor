package dao

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()

	result := m.Run()
	dao.PrepareTestData([]string{
		`DELETE FROM "retention_filter_metadata"; ALTER SEQUENCE retention_filter_metadata_id_seq RESTART;`,
		`DELETE FROM "retention_policy"; ALTER SEQUENCE retention_policy_id_seq RESTART;`,
		`DELETE FROM repository; ALTER SEQUENCE repository_repository_id_seq RESTART;`,
	}, nil)

	os.Exit(result)
}
