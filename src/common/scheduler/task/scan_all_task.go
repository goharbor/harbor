package task

import (
	"github.com/goharbor/harbor/src/ui/utils"
)

//ScanAllTask is task of scanning all tags.
type ScanAllTask struct{}

//NewScanAllTask is constructor of creating ScanAllTask.
func NewScanAllTask() *ScanAllTask {
	return &ScanAllTask{}
}

//Name returns the name of the task.
func (sat *ScanAllTask) Name() string {
	return "scan all"
}

//Run the actions.
func (sat *ScanAllTask) Run() error {
	return utils.ScanAllImages()
}
