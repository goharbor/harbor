package api

import (
	"errors"
	"fmt"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/label/model"
)

const (
	versionParam = ":version"
	idParam      = ":id"
)

// ChartLabelAPI handles the requests of marking/removing labels to/from charts.
type ChartLabelAPI struct {
	LabelResourceAPI
	project       *proModels.Project
	chartFullName string
}

// Prepare required material for follow-up actions.
func (cla *ChartLabelAPI) Prepare() {
	// Super
	cla.LabelResourceAPI.Prepare()

	// Check authorization
	if !cla.SecurityCtx.IsAuthenticated() {
		cla.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	project := cla.GetStringFromPath(namespaceParam)

	// Project should be a valid existing one
	existingProject, err := cla.ProjectCtl.Get(cla.Context(), project)
	if err != nil {
		cla.SendError(err)
		return
	}
	cla.project = existingProject

	// Check the existence of target chart
	chartName := cla.GetStringFromPath(nameParam)
	version := cla.GetStringFromPath(versionParam)

	if _, err = chartController.GetChartVersion(project, chartName, version); err != nil {
		cla.SendNotFoundError(err)
		return
	}
	cla.chartFullName = fmt.Sprintf("%s/%s:%s", project, chartName, version)
}

func (cla *ChartLabelAPI) requireAccess(action rbac.Action) bool {
	return cla.RequireProjectAccess(cla.project.ProjectID, action, rbac.ResourceHelmChartVersionLabel)
}

// MarkLabel handles the request of marking label to chart.
func (cla *ChartLabelAPI) MarkLabel() {
	if !cla.requireAccess(rbac.ActionCreate) {
		return
	}

	l := &model.Label{}
	if err := cla.DecodeJSONReq(l); err != nil {
		cla.SendBadRequestError(err)
		return
	}

	label, ok := cla.validate(l.ID, cla.project.ProjectID)
	if !ok {
		return
	}

	label2Res := &models.ResourceLabel{
		LabelID:      label.ID,
		ResourceType: common.ResourceTypeChart,
		ResourceName: cla.chartFullName,
	}

	cla.markLabelToResource(label2Res)
}

// RemoveLabel handles the request of removing label from chart.
func (cla *ChartLabelAPI) RemoveLabel() {
	if !cla.requireAccess(rbac.ActionDelete) {
		return
	}

	lID, err := cla.GetInt64FromPath(idParam)
	if err != nil {
		cla.SendInternalServerError(err)
		return
	}

	label, ok := cla.exists(lID)
	if !ok {
		return
	}

	cla.removeLabelFromResource(common.ResourceTypeChart, cla.chartFullName, label.ID)
}

// GetLabels gets labels for the specified chart version.
func (cla *ChartLabelAPI) GetLabels() {
	cla.getLabelsOfResource(common.ResourceTypeChart, cla.chartFullName)
}
