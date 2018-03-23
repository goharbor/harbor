package task

import (
	"sync"
)

//Store is designed to keep the tasks.
type Store interface {
	//GetTasks return the current existing list in store.
	GetTasks() []Task

	//AddTasks is used to append tasks to the list.
	AddTasks(tasks ...Task)
}

//DefaultStore is the default implemetation of Store interface.
type DefaultStore struct {
	//To sync the related operations.
	*sync.RWMutex

	//The space to keep the tasks.
	tasks []Task
}

//NewDefaultStore is constructor method for DefaultStore.
func NewDefaultStore() *DefaultStore {
	return &DefaultStore{new(sync.RWMutex), []Task{}}
}

//GetTasks implements the same method in Store interface.
func (ds *DefaultStore) GetTasks() []Task {
	copyList := []Task{}

	ds.RLock()
	defer ds.RUnlock()

	if ds.tasks != nil && len(ds.tasks) > 0 {
		copyList = append(copyList, ds.tasks...)
	}

	return copyList
}

//AddTasks implements the same method in Store interface.
func (ds *DefaultStore) AddTasks(tasks ...Task) {
	//Double confirm.
	if ds.tasks == nil {
		ds.tasks = []Task{}
	}

	ds.Lock()
	defer ds.Unlock()

	ds.tasks = append(ds.tasks, tasks...)
}
