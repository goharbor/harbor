package api

import (
	pkg_label "github.com/goharbor/harbor/src/pkg/label"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/label"
	"github.com/goharbor/harbor/src/pkg/label/model"
)

// LabelResourceAPI provides the related basic functions to handle marking labels to resources
type LabelResourceAPI struct {
	BaseController
	labelManager label.Manager
}

// Prepare resources for follow-up actions.
func (lra *LabelResourceAPI) Prepare() {
	lra.BaseController.Prepare()

	// Create label manager
	lra.labelManager = &label.BaseManager{
		LabelMgr: pkg_label.Mgr,
	}
}

func (lra *LabelResourceAPI) getLabelsOfResource(rType string, rIDOrName interface{}) {
	labels, err := lra.labelManager.GetLabelsOfResource(rType, rIDOrName)
	if err != nil {
		lra.handleErrors(err)
		return
	}

	lra.Data["json"] = labels
	lra.ServeJSON()
}

func (lra *LabelResourceAPI) markLabelToResource(rl *models.ResourceLabel) {
	labelID, err := lra.labelManager.MarkLabelToResource(rl)
	if err != nil {
		lra.handleErrors(err)
		return
	}

	// return the ID of label and return status code 200 rather than 201 as the label is not created
	lra.Redirect(http.StatusOK, strconv.FormatInt(labelID, 10))
}

func (lra *LabelResourceAPI) removeLabelFromResource(rType string, rIDOrName interface{}, labelID int64) {
	if err := lra.labelManager.RemoveLabelFromResource(rType, rIDOrName, labelID); err != nil {
		lra.handleErrors(err)
		return
	}
}

// eat the error of validate method of label manager
func (lra *LabelResourceAPI) validate(labelID, projectID int64) (*model.Label, bool) {
	label, err := lra.labelManager.Validate(labelID, projectID)
	if err != nil {
		lra.handleErrors(err)
		return nil, false
	}

	return label, true
}

// eat the error of exists method of label manager
func (lra *LabelResourceAPI) exists(labelID int64) (*model.Label, bool) {
	label, err := lra.labelManager.Exists(labelID)
	if err != nil {
		return nil, false
	}

	return label, true
}

// Handle different kinds of errors.
func (lra *LabelResourceAPI) handleErrors(err error) {
	switch err.(type) {
	case *label.ErrLabelBadRequest:
		lra.SendBadRequestError(err)
	case *label.ErrLabelConflict:
		lra.SendConflictError(err)
	case *label.ErrLabelNotFound:
		lra.SendNotFoundError(err)
	default:
		lra.SendInternalServerError(err)
	}
}
