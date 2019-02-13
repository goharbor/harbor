package schedule

import (
	"testing"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

var replicator *DefaultReplicator

func init() {
	replicator = NewDefaultReplicator(TestClient{})
}

type TestClient struct {
}

func (client TestClient) SubmitJob(*models.JobData) (string, error) {
	return "submited-uuid", nil
}
func (client TestClient) GetJobLog(uuid string) ([]byte, error) {
	return []byte("job log"), nil
}
func (client TestClient) PostAction(uuid, action string) error {
	return nil
}

func TestDefaultReplicator_Schedule(t *testing.T) {
	tasks, err := generateData()
	if err != nil {
		t.Error(err)
	}
	for _, task := range tasks {
		t.Log(*task)
	}

}

//func TestDefaultReplicator_SubmitJobs(t *testing.T) {
//	config.Init()
//	tasks, err := generateData()
//	if err != nil {
//		t.Error(err)
//	}
//	for _, task := range tasks {
//		task.ID = 22
//	}
//	newTasks, newErr := replicator.SubmitTasks(tasks)
//	if newErr != nil {
//		t.Error(newErr)
//	}
//	for _, task := range newTasks {
//		t.Log(*task)
//	}
//}

func TestDefaultReplicator_StopExecution(t *testing.T) {
	err := replicator.StopExecution("id")
	if err != nil {
		t.Error(err)
	}
}

func generateData() ([]*model.Task, error) {
	srcResource := &model.Resource{
		Metadata: &model.ResourceMetadata{
			Namespace: &model.Namespace{
				Metadata: map[string]interface{}{
					"resource":     "1",
					"dst_registry": "1",
					"namespace":    "1",
				},
			},
			Vtags:  []string{"latest"},
			Labels: []string{"latest"},
		},
		Registry: &model.Registry{
			Credential: &model.Credential{},
		},
	}
	destResource := &model.Resource{
		Metadata: &model.ResourceMetadata{
			Namespace: &model.Namespace{
				Metadata: map[string]interface{}{
					"resource":     "2",
					"dst_registry": "2",
					"namespace":    "2",
				},
			},
			Vtags:  []string{"latest"},
			Labels: []string{"latest"},
		},
		Registry: &model.Registry{
			Credential: &model.Credential{},
		},
	}
	tasks, err := replicator.Schedule([]*model.Resource{srcResource}, []*model.Resource{destResource})
	return tasks, err
}
