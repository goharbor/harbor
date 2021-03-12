package nydus

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/google/uuid"
	"os"
	"strings"
)

func init() {
	// keep only the latest created 50 nydus execution records
	task.SetExecutionSweeperCount(NydusVendorType, 50)
}

const (
	// NydusVendorType ...
	NydusVendorType = "NYDUS"
)

// DefaultController is a default global nydus controller
var DefaultController = NewController()

type Controller interface {
	// Convert images to nydus image
	Convert(ctx context.Context, artifact *artifact.Artifact, trigger string) error
	// Stop stop a nydus convert job
	Stop(ctx context.Context, id int64) error
	// GetTaskLog gets log of the specific task
	GetTaskLog(ctx context.Context, id int64) ([]byte, error)
}

// controller ...
type controller struct {
	rc      robot.Controller
	taskMgr task.Manager
	exeMgr  task.ExecutionManager
}

// NewController ...
func NewController() Controller {
	return &controller{
		rc:      robot.Ctl,
		exeMgr:  task.ExecMgr,
		taskMgr: task.Mgr,
	}
}

func (bc *controller) Convert(ctx context.Context, artifact *artifact.Artifact, trigger string) error {
	r, err := bc.makeRobotAccount(ctx, artifact.ProjectID, artifact.RepositoryName)
	if err != nil {
		return err
	}

	if len(artifact.Tags) <= 0 {
		return errors.New("the tag of artifact is empty")
	}

	para := make(map[string]interface{})
	para["core_url"] = os.Getenv("CORE_URL")
	para["username"] = r.Name
	para["password"] = r.Secret
	para["repository"] = artifact.RepositoryName
	para["tag"] = artifact.Tags[0].Name

	execID, err := bc.exeMgr.Create(ctx, NydusVendorType, -1, trigger, para)
	if err != nil {
		return err
	}
	_, err = bc.taskMgr.Create(ctx, execID, &task.Job{
		Name: job.Nydus,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: para,
	})
	if err != nil {
		return err
	}

	return nil
}

// Stop ...
func (bc *controller) Stop(ctx context.Context, id int64) error {
	return bc.exeMgr.Stop(ctx, id)
}

// GetTask ...
func (bc *controller) GetTask(ctx context.Context, id int64) (*Task, error) {
	tasks, err := bc.taskMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ID":         id,
			"VendorType": NydusVendorType,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessage("nydus task %d not found", id)
	}
	return convertTask(tasks[0]), nil
}

func (bc *controller) GetTaskLog(ctx context.Context, id int64) ([]byte, error) {
	_, err := bc.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return bc.taskMgr.GetLog(ctx, id)
}

// makeRobotAccount creates a robot account based on the arguments for convert.
func (bc *controller) makeRobotAccount(ctx context.Context, projectID int64, repository string) (*robot.Robot, error) {
	UUID, err := uuid.NewUUID()
	if err != nil {
		return nil, errors.Wrap(err, "convert controller: make robot account")
	}

	projectName := strings.Split(repository, "/")[0]

	robotReq := &robot.Robot{
		Robot: model.Robot{
			Name:        fmt.Sprintf("nydus-auto-%s", UUID),
			Description: "for scan",
			ProjectID:   projectID,
		},
		Level: robot.LEVELPROJECT,
		Permissions: []*robot.Permission{
			{
				Kind:      "project",
				Namespace: projectName,
				Access: []*types.Policy{
					{
						Resource: rbac.ResourceRepository,
						Action:   rbac.ActionPull,
					},
					{
						Resource: rbac.ResourceRepository,
						Action:   rbac.ActionPush,
					},
					{
						Resource: rbac.ResourceRepository,
						Action:   rbac.ActionScannerPull,
					},
				},
			},
		},
	}

	rb, pwd, err := bc.rc.Create(ctx, robotReq)
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}

	r, err := bc.rc.Get(ctx, rb, &robot.Option{WithPermission: false})
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}
	r.Secret = pwd
	return r, nil
}

func convertTask(task *task.Task) *Task {
	return &Task{
		ID:            task.ID,
		ExecutionID:   task.ExecutionID,
		Status:        task.Status,
		StatusMessage: task.StatusMessage,
		RunCount:      task.RunCount,
		Repository:    task.GetStringFromExtraAttrs("repository"),
		Tag:           task.GetStringFromExtraAttrs("tag"),
		JobID:         task.JobID,
		CreationTime:  task.CreationTime,
		StartTime:     task.StartTime,
		UpdateTime:    task.UpdateTime,
		EndTime:       task.EndTime,
	}
}
