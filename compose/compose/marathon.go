package compose

import (
	"fmt"
	"github.com/vmware/harbor/utils"
)

const (
	DEFAULT_CPU       = 0.2
	DEFAULT_MEM       = 2
	DEFAULT_INSTANCES = 2
)

// mesos marathon specfic config
type MarathonConfig struct {
	ClusterId    int32    `json:"clusterId" yaml:"clusterId"`
	AppName      string   `json:"appName" yaml:"appName"`
	ImageVersion string   `json:"imageVersion" yaml:"imageVersion"`
	Cpu          float32  `json:"cpu" yaml:"cpu"`
	Mem          float32  `json:"mem" yaml:"mem"`
	Instances    int32    `json:"instances" yaml:"instances"`
	LogPaths     []string `json:"log_paths" yaml:"log_paths"`
}

func (self *MarathonConfig) Defaultlize() {
	if utils.FloatEquals(self.Cpu, 0.0) {
		self.Cpu = DEFAULT_CPU
	}

	if utils.FloatEquals(self.Mem, 0.0) {
		self.Mem = DEFAULT_MEM
	}

	if self.Instances == 0 {
		self.Instances = DEFAULT_INSTANCES
	}

	self.LogPaths = []string{}

}

func (mc *MarathonConfig) ToString() string {
	mcBasic := ""
	mcBasic = "\n"
	mcBasic += fmt.Sprintf("ClusterId: %-30d\n", mc.ClusterId)
	mcBasic += fmt.Sprintf("AppName: %-30s\n", mc.AppName)
	mcBasic += fmt.Sprintf("Cpu: %-30f\n", mc.Cpu)
	mcBasic += fmt.Sprintf("Mem: %-30f\n", mc.Mem)
	mcBasic += fmt.Sprintf("Instances: %-30d\n", mc.Instances)

	return mcBasic
}
