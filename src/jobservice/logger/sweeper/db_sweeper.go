package sweeper

import (
	"time"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
)

var dbInit = make(chan int, 1)

// DBSweeper is used to sweep the DB logs
type DBSweeper struct {
	duration int
}

// NewDBSweeper is constructor of DBSweeper
func NewDBSweeper(duration int) *DBSweeper {
	return &DBSweeper{
		duration: duration,
	}
}

// Sweep logs
func (dbs *DBSweeper) Sweep() (int, error) {
	// DB initialization not completed, waiting
	<-dbInit

	// Start to sweep logs
	before := time.Now().Add(time.Duration(dbs.duration) * oneDay * -1)
	count, err := dao.DeleteJobLogsBefore(before)

	if err != nil {
		return 0, fmt.Errorf("sweep logs in DB failed before %s with error: %s", before, err)
	}

	return int(count), nil
}

// Duration for sweeping
func (dbs *DBSweeper) Duration() int {
	return dbs.duration
}

// prepare sweeping
func PrepareDBSweep() error {
	dbInit <- 1
	return nil
}