package job

import (
	"fmt"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

const (
	INFO = "info"
	WARN = "warning"
	ERR  = "error"
)

type Logger struct {
	ID int64
}

func (l *Logger) Infof(format string, v ...interface{}) {
	err := dao.AddJobLog(l.ID, INFO, fmt.Sprintf(format, v...))
	if err != nil {
		log.Warningf("Failed to add job log, id: %d, error: %v", l.ID, err)
	}
}

func (l *Logger) Warningf(format string, v ...interface{}) {
	err := dao.AddJobLog(l.ID, WARN, fmt.Sprintf(format, v...))
	if err != nil {
		log.Warningf("Failed to add job log, id: %d, error: %v", l.ID, err)
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	err := dao.AddJobLog(l.ID, ERR, fmt.Sprintf(format, v...))
	if err != nil {
		log.Warningf("Failed to add job log, id: %d, error: %v", l.ID, err)
	}
}
