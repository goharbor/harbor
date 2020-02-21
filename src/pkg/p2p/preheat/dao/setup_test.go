package dao

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}
