// Copyright 2018 The Harbor Authors. All rights reserved.

package logger

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const (
	oneDay = 3600 * 24
)

//Sweeper takes charge of archive the outdated log files of jobs.
type Sweeper struct {
	context context.Context
	workDir string
	period  uint
}

//NewSweeper creates new prt of Sweeper
func NewSweeper(ctx context.Context, workDir string, period uint) *Sweeper {
	return &Sweeper{ctx, workDir, period}
}

//Start to work
func (s *Sweeper) Start() {
	go s.loop()
	Info("Logger sweeper is started")
}

func (s *Sweeper) loop() {
	//Apply default if needed before starting
	if s.period == 0 {
		s.period = 1
	}

	defer func() {
		Info("Logger sweeper is stopped")
	}()

	//First run
	go s.clear()

	ticker := time.NewTicker(time.Duration(s.period*oneDay+5) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.context.Done():
			return
		case <-ticker.C:
			go s.clear()
		}
	}
}

func (s *Sweeper) clear() {
	var (
		cleared uint
		count   = &cleared
	)

	Info("Start to clear the job outdated log files")
	defer func() {
		Infof("%d job outdated log files cleared", *count)
	}()

	logFiles, err := ioutil.ReadDir(s.workDir)
	if err != nil {
		Errorf("Failed to get the outdated log files under '%s' with error: %s\n", s.workDir, err)
		return
	}
	if len(logFiles) == 0 {
		return
	}

	for _, logFile := range logFiles {
		if logFile.ModTime().Add(time.Duration(s.period*oneDay) * time.Second).Before(time.Now()) {
			logFilePath := fmt.Sprintf("%s%s%s", s.workDir, string(os.PathSeparator), logFile.Name())
			if err := os.Remove(logFilePath); err == nil {
				cleared++
			} else {
				Warningf("Failed to remove log file '%s'\n", logFilePath)
			}
		}
	}
}
