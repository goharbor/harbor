package task

//Task is used to synchronously run specific action(s).
type Task interface {
	//Name should return the name of the task.
	Name() string

	//Run the concrete code here
	Run() error
}
