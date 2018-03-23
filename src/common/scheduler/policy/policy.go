package policy

import (
	"github.com/vmware/harbor/src/common/scheduler/task"
)

//Policy is an if-then logic to determine how the attached tasks should be
//executed based on the evaluation result of the defined conditions.
//E.g:
//  Daily execute TASK between 2017/06/24 and 2018/06/23
//  Execute TASK at 2017/09/01 14:30:00
//
//Each policy should have a name to identify itself.
//Please be aware that policy with no tasks will be treated as invalid.
//
type Policy interface {
	//Name will return the name of the policy.
	//If the policy supports multiple instances, please make sure the name is unique as an UUID.
	Name() string

	//Tasks will return the attached tasks with this policy.
	Tasks() []task.Task

	//AttachTasks is to attach tasks to this policy
	AttachTasks(...task.Task) error

	//Done will setup a channel for other components to check whether or not
	//the policy is completed. Possibly designed for the none loop policy.
	Done() <-chan bool

	//Evaluate the policy based on its definition and return the result via
	//result channel. Policy is enabled after it is evaluated.
	//Make sure Evaluate is idempotent, that means one policy can be only enabled
	//only once even if Evaluate is called more than one times.
	Evaluate() (<-chan bool, error)

	//Disable the enabled policy and release all the allocated resources.
	Disable() error

	//Equal will compare the two policies based on related factors if existing such as confgiuration etc.
	//to determine whether the two policies are same ones or not. Please pay attention that, not every policy
	//needs to support this method. If no need, please directly return false to indicate each policies are
	//different.
	Equal(p Policy) bool

	//IsEnabled is to indicate whether the policy is enabled or not (disabled).
	IsEnabled() bool
}
