package getter

import (
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/joblog"

	"github.com/goharbor/harbor/src/jobservice/errs"
)

// DBGetter is responsible for retrieving DB log data
type DBGetter struct {
}

// NewDBGetter is constructor of DBGetter
func NewDBGetter() *DBGetter {
	return &DBGetter{}
}

// Retrieve implements @Interface.Retrieve
func (dbg *DBGetter) Retrieve(logID string) ([]byte, error) {
	if len(logID) == 0 {
		return nil, errors.New("empty log identify")
	}

	jobLog, err := joblog.Mgr.Get(orm.Context(), logID)
	if err != nil {
		// Other errors have been ignored by GetJobLog()
		return nil, errs.NoObjectFoundError(fmt.Sprintf("log entity: %s", logID))
	}

	return []byte(jobLog.Content), nil
}
