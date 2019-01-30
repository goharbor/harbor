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
	tasks, err := replicator.Schedule([]*model.Resource{srcResource}, []*model.Resource{destResource})
	if err != nil {
		t.Error(err)
	}
	for _, task := range tasks {
		t.Log(*task)
	}

}

func TestDefaultReplicator_StopTransfer(t *testing.T) {
	err := replicator.StopTransfer("job_id")
	if err != nil {
		t.Error(err)
	}
}
