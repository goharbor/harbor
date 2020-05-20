package getter

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/logger/backend"
	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
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

// TestDBGetter
func TestDBGetter(t *testing.T) {
	uuid := "uuid_for_unit_test_getter"
	l, err := backend.NewDBLogger(uuid, "DEBUG", 4)
	require.Nil(t, err)

	l.Debug("JobLog Debug: TestDBLoggerGetter")
	err = l.Close()
	require.NoError(t, err)

	dbGetter := NewDBGetter()
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

// TestDBGetterError
func TestDBGetterError(t *testing.T) {
	uuid := "uuid_for_unit_test_getter_error"
	l, err := backend.NewDBLogger(uuid, "DEBUG", 4)
	require.Nil(t, err)

	l.Debug("JobLog Debug: TestDBLoggerGetter")
	err = l.Close()
	require.NoError(t, err)

	dbGetter := NewDBGetter()
	_, err = dbGetter.Retrieve("")
	require.NotNil(t, err)
	_, err = dbGetter.Retrieve("not_exist_uuid")
	require.NotNil(t, err)

	err = sweeper.PrepareDBSweep()
	require.NoError(t, err)
	dbSweeper := sweeper.NewDBSweeper(-1)
	count, err := dbSweeper.Sweep()
	require.Nil(t, err)
	require.Equal(t, 1, count)
}
