package trigger

import (
	"fmt"
	"net/http"

	"github.com/vmware/harbor/src/common/dao"
	common_http "github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/job"
	job_models "github.com/vmware/harbor/src/common/job/models"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/utils"
)

//ScheduleTrigger will schedule a alternate policy to provide 'daily' and 'weekly' trigger ways.
type ScheduleTrigger struct {
	params ScheduleParam
}

//NewScheduleTrigger is constructor of ScheduleTrigger
func NewScheduleTrigger(params ScheduleParam) *ScheduleTrigger {
	return &ScheduleTrigger{
		params: params,
	}
}

//Kind is the implementation of same method defined in Trigger interface
func (st *ScheduleTrigger) Kind() string {
	return replication.TriggerKindSchedule
}

//Setup is the implementation of same method defined in Trigger interface
func (st *ScheduleTrigger) Setup() error {
	metadata := &job_models.JobMetadata{
		JobKind: job.JobKindPeriodic,
	}
	switch st.params.Type {
	case replication.TriggerScheduleDaily:
		h, m, s := parseOfftime(st.params.Offtime)
		metadata.Cron = fmt.Sprintf("%d %d %d * * *", s, m, h)
	case replication.TriggerScheduleWeekly:
		h, m, s := parseOfftime(st.params.Offtime)
		metadata.Cron = fmt.Sprintf("%d %d %d * * %d", s, m, h, st.params.Weekday%7)
	default:
		return fmt.Errorf("unsupported schedual trigger type: %s", st.params.Type)
	}

	id, err := dao.AddRepJob(models.RepJob{
		Repository: "N/A",
		PolicyID:   st.params.PolicyID,
		Operation:  models.RepOpSchedule,
	})
	if err != nil {
		return err
	}
	uuid, err := utils.GetJobServiceClient().SubmitJob(&job_models.JobData{
		Name: job.ImageReplicate,
		Parameters: map[string]interface{}{
			"policy_id": st.params.PolicyID,
			"url":       config.InternalUIURL(),
			"insecure":  true,
		},
		Metadata: metadata,
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/replication/%d",
			config.InternalUIURL(), id),
	})
	if err != nil {
		// clean up the job record in database
		if e := dao.DeleteRepJob(id); e != nil {
			log.Errorf("failed to delete job %d: %v", id, e)
		}
		return err
	}
	return dao.SetRepJobUUID(id, uuid)
}

//Unset is the implementation of same method defined in Trigger interface
func (st *ScheduleTrigger) Unset() error {
	jobs, err := dao.GetRepJobs(&models.RepJobQuery{
		PolicyID:   st.params.PolicyID,
		Operations: []string{models.RepOpSchedule},
	})
	if err != nil {
		return err
	}
	if len(jobs) != 1 {
		log.Warningf("only one job should be found, but found %d now", len(jobs))
	}

	for _, j := range jobs {
		if err = utils.GetJobServiceClient().PostAction(j.UUID, job.JobActionStop); err != nil {
			// if the job specified by UUID is not found in jobservice, delete the job
			// record from database
			if e, ok := err.(*common_http.Error); !ok || e.Code != http.StatusNotFound {
				return err
			}
		}
		if err = dao.DeleteRepJob(j.ID); err != nil {
			return err
		}
	}
	return nil
}

func parseOfftime(offtime int64) (hour, minite, second int) {
	offtime = offtime % (3600 * 24)
	hour = int(offtime / 3600)
	offtime = offtime % 3600
	minite = int(offtime / 60)
	second = int(offtime % 60)
	return
}
