package policy

import "github.com/vmware/harbor/src/common/scheduler/task"

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
	//Return the name of the policy.
	Name() string

	//Return the attached tasks with this policy.
	Tasks() []task.Task

	//Attach tasks to this policy
	AttachTasks(...task.Task) error

	//Done will setup a channel for other components to check whether or not
	//the policy is completed. Possibly designed for the none loop policy.
	Done() chan bool

	//Evaluate the policy based on its definition and return the result via
	//result channel. Policy is enabled after it is evaluated.
	//Make sure Evaluate is idempotent, that means one policy can be only enabled
	//only once even if Evaluate is called more than one times.
	Evaluate() chan EvaluationResult

	//Disable the enabled policy and release all the allocated resources.
	//Disable should also send signal to the terminated channel which returned by Done.
	Disable() error
}
