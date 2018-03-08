package core

import (
	"github.com/vmware/harbor/src/jobservice_v2/models"
)

//Controller implement the core interface and provides related job handle methods.
//Controller will coordinate the lower components to complete the process as a commander role.
type Controller struct{}

//NewController is constructor of Controller.
func NewController() *Controller {
	return &Controller{}
}

//LaunchJob is implementation of same method in core interface.
func (c *Controller) LaunchJob(ctx BaseContext, req models.JobRequest) (models.JobStats, error) {
	return models.JobStats{
		JobID: "111112222xxx",
	}, nil
}

//GetJob is implementation of same method in core interface.
func (c *Controller) GetJob(jobID string) (models.JobStats, error) {
	return models.JobStats{}, nil
}

//StopJob is implementation of same method in core interface.
func (c *Controller) StopJob(jobID string) error {
	return nil
}

//RetryJob is implementation of same method in core interface.
func (c *Controller) RetryJob(ctx BaseContext, jonID string) error {
	return nil
}

//CheckStatus is implementation of same method in core interface.
func (c *Controller) CheckStatus() (models.JobServiceStats, error) {
	return models.JobServiceStats{}, nil
}
