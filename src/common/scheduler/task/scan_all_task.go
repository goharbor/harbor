package task

import "github.com/vmware/harbor/src/ui/utils"

//ScanAllTask is task of scanning all tags.
type ScanAllTask struct{}

//NewScanAllTask is constructor of creating ScanAllTask.
func NewScanAllTask() *ScanAllTask {
	return &ScanAllTask{}
}

//TaskName returns the name of the task.
func (sat *ScanAllTask) TaskName() string {
	return "scan all"
}

//Run the actions.
func (sat *ScanAllTask) Run() error {
	return utils.ScanAllImages()
}
