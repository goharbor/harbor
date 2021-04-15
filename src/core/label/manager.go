package label

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/label"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/label/model"
)

// Manager defines the related operations for label management
type Manager interface {
	// Mark label to the resource.
	//
	// If succeed, the relationship ID will be returned.
	// Otherwise, an non-nil error will be returned.
	MarkLabelToResource(label *models.ResourceLabel) (int64, error)

	// Remove the label from the resource.
	// Resource type and ID(/name) should be provided to identify the relationship.
	//
	// An non-nil error will be got if meet any issues or nil error returned.
	RemoveLabelFromResource(resourceType string, resourceIDOrName interface{}, labelID int64) error

	// Get labels for the specified resource.
	// Resource is identified by the resource type and ID(/name).
	//
	// If succeed, a label list is returned.
	// Otherwise, a non-nil error will be returned.
	GetLabelsOfResource(resourceType string, resourceIDOrName interface{}) ([]*model.Label, error)

	// Check the existence of the specified label.
	//
	// If label existing, a non-nil label object is returned and nil error is set.
	// A non-nil error will be set if any issues met while checking or label is not found.
	Exists(labelID int64) (*model.Label, error)

	// Validate if the scope of the input label is correct.
	// If the scope is project level, the projectID is required then.
	//
	// If everything is ok, an validated label reference will be returned.
	// Otherwise, a non-nil error is returned.
	Validate(labelID int64, projectID int64) (*model.Label, error)
}

// BaseManager is the default implementation of the Manager interface.
type BaseManager struct {
	LabelMgr label.Manager
}

// MarkLabelToResource is the implementation of same method in Manager interface.
func (bm *BaseManager) MarkLabelToResource(label *models.ResourceLabel) (int64, error) {
	if label == nil {
		return -1, errors.New("nil label object")
	}

	// Use ID or name of resource. ID first.
	var rIDOrName interface{}
	if label.ResourceID != 0 {
		rIDOrName = label.ResourceID
	} else {
		rIDOrName = label.ResourceName
	}

	rlabel, err := dao.GetResourceLabel(label.ResourceType, rIDOrName, label.LabelID)
	if err != nil {
		return -1, fmt.Errorf("failed to check the existence of label %d for resource %s %v: %v", label.LabelID, label.ResourceType, rIDOrName, err)
	}

	if rlabel != nil {
		return -1, NewErrLabelConflict(label.LabelID, label.ResourceType, rIDOrName)
	}

	if _, err := dao.AddResourceLabel(label); err != nil {
		return -1, fmt.Errorf("failed to add label %d to resource %s %v: %v", label.LabelID, label.ResourceType, rIDOrName, err)
	}

	// return the ID of label
	return label.LabelID, nil
}

// RemoveLabelFromResource is the implementation of same method in Manager interface.
func (bm *BaseManager) RemoveLabelFromResource(resourceType string, resourceIDOrName interface{}, labelID int64) error {
	rl, err := dao.GetResourceLabel(resourceType, resourceIDOrName, labelID)
	if err != nil {
		return fmt.Errorf("failed to check the existence of label %d for resource %s %v: %v", labelID, resourceType, resourceIDOrName, err)
	}

	if rl == nil {
		return NewErrLabelNotFound(labelID, resourceType, resourceIDOrName)
	}

	if err = dao.DeleteResourceLabel(rl.ID); err != nil {
		return fmt.Errorf("failed to delete resource label record %d: %v", rl.ID, err)
	}

	return nil
}

// GetLabelsOfResource is the implementation of same method in Manager interface.
func (bm *BaseManager) GetLabelsOfResource(resourceType string, resourceIDOrName interface{}) ([]*model.Label, error) {
	labels, err := dao.GetLabelsOfResource(resourceType, resourceIDOrName)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels of resource %s %v: %v", resourceType, resourceIDOrName, err)
	}

	return labels, nil
}

// Exists is the implementation of same method in Manager interface.
func (bm *BaseManager) Exists(labelID int64) (*model.Label, error) {
	label, err := bm.LabelMgr.Get(orm.Context(), labelID)
	if err != nil {
		if errors.IsErr(err, errors.NotFoundCode) {
			return nil, NewErrLabelNotFound(labelID, "", nil)
		}
		return nil, fmt.Errorf("failed to get label %d: %v", labelID, err)
	}

	return label, nil
}

// Validate is the implementation of same method in Manager interface.
func (bm *BaseManager) Validate(labelID int64, projectID int64) (*model.Label, error) {
	label, err := bm.LabelMgr.Get(orm.Context(), labelID)
	if err != nil {
		if errors.IsErr(err, errors.NotFoundCode) {
			return nil, NewErrLabelNotFound(labelID, "", nil)
		}
		return nil, fmt.Errorf("failed to get label %d: %v", labelID, err)
	}

	if label.Level != common.LabelLevelUser {
		return nil, NewErrLabelBadRequest("only user level labels can be used")
	}

	if label.Scope == common.LabelScopeProject {
		if projectID != label.ProjectID {
			return nil, NewErrLabelBadRequest("can not add labels which don't belong to the project to the resources under the project")
		}
	}

	return label, nil
}
