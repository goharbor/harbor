package scheduler

import (
	"encoding/json"
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

func TestDefaultReplicator_Preprocess(t *testing.T) {
	items, err := generateData()
	if err != nil {
		t.Error(err)
	}
	for _, item := range items {
		content, err := json.Marshal(item)
		if err != nil {
			t.Error(err)
		}
		t.Log(string(content))
	}

}

//func TestDefaultReplicator_Schedule(t *testing.T) {
//	//	config.Init()
//	items, err := generateData()
//	if err != nil {
//		t.Error(err)
//	}
//	for _, item := range items {
//		item.TaskID = 22
//	}
//	results, newErr := replicator.Schedule(items)
//	if newErr != nil {
//		t.Error(newErr)
//	}
//	for _, result := range results {
//		content, err := json.Marshal(result)
//		if err != nil {
//			t.Error(err)
//		}
//		t.Log(string(content))
//	}
//}
//
func TestDefaultReplicator_Stop(t *testing.T) {
	err := replicator.Stop("id")
	if err != nil {
		t.Error(err)
	}
}

func generateData() ([]*ScheduleItem, error) {
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
	items, err := replicator.Preprocess([]*model.Resource{srcResource}, []*model.Resource{destResource})
	return items, err
}
