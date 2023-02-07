package getter

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/joblog"
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

	sz := int64(len(jobLog.Content))
	var buf []byte
	sizeLimit := logSizeLimit()
	if sizeLimit <= 0 {
		buf = []byte(jobLog.Content)
		return buf, nil
	}
	if sz > sizeLimit {
		buf = []byte(jobLog.Content[sz-sizeLimit:])
	}
	return buf, nil
}
