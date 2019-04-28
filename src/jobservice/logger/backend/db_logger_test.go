package backend

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/jobservice/logger/getter"
	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	// databases := []string{"mysql", "sqlite"}
	databases := []string{"postgresql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)

		result := 1
		switch database {
		case "postgresql":
			dao.PrepareTestForPostgresSQL()
		default:
			log.Fatalf("invalid database: %s", database)
		}

		result = m.Run()

		if result != 0 {
			os.Exit(result)
		}
	}

}

// Test DB logger
func TestDBLogger(t *testing.T) {
	uuid := "uuid_for_unit_test"
	l, err := NewDBLogger(uuid, "DEBUG", 4)
	require.Nil(t, err)

	l.Debug("JobLog Debug: TestDBLogger")
	l.Info("JobLog Info: TestDBLogger")
	l.Warning("JobLog Warning: TestDBLogger")
	l.Error("JobLog Error: TestDBLogger")
	l.Debugf("JobLog Debugf: %s", "TestDBLogger")
	l.Infof("JobLog Infof: %s", "TestDBLogger")
	l.Warningf("JobLog Warningf: %s", "TestDBLogger")
	l.Errorf("JobLog Errorf: %s", "TestDBLogger")

	_ = l.Close()

	dbGetter := getter.NewDBGetter()
	ll, err := dbGetter.Retrieve(uuid)
	require.Nil(t, err)
	log.Infof("get logger %s", ll)

	err = sweeper.PrepareDBSweep()
	require.NoError(t, err)
	dbSweeper := sweeper.NewDBSweeper(-1)
	count, err := dbSweeper.Sweep()
	require.Nil(t, err)
	require.Equal(t, 1, count)
}
