package logger

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/logger/backend"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/require"
	"os"
	"path"
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

// TestEntry
func TestEntry(t *testing.T) {
	var loggers = make([]Interface, 0)
	uuid := "uuid_for_unit_test"
	dbl, err := backend.NewDBLogger(uuid, "DEBUG", 4)
	require.Nil(t, err)
	loggers = append(loggers, dbl)

	fl, err := backend.NewFileLogger("DEBUG", path.Join(os.TempDir(), "TestFileLogger.log"), 4)
	require.Nil(t, err)
	loggers = append(loggers, fl)

	en := NewEntry(loggers)

	en.Debug("JobLog Debug: TestEntry")
	en.Info("JobLog Info: TestEntry")
	en.Warning("JobLog Warning: TestEntry")
	en.Error("JobLog Error: TestEntry")
	en.Debugf("JobLog Debugf: %s", "TestEntry")
	en.Infof("JobLog Infof: %s", "TestEntry")
	en.Warningf("JobLog Warningf: %s", "TestEntry")
	en.Errorf("JobLog Errorf: %s", "TestEntry")

	err = en.Close()
	require.Nil(t, err)
}
