package task

//Task is used to synchronously run specific action(s).
type Task interface {
	//Name of the task.
	TaskName() string

	//Run the concrete code here
	Run() error
}
