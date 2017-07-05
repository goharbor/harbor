package task

//Task is used to synchronously run specific action(s).
type Task interface {
	//TaskName should return the name of the task.
	TaskName() string

	//Run the concrete code here
	Run() error
}
