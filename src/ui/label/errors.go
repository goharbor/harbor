package label

import (
	"fmt"
)

// ErrLabelBase contains the basic required info for building the final errors.
type ErrLabelBase struct {
	LabelID          int64
	ResourceType     string
	ResourceIDOrName interface{}
}

// ErrLabelNotFound defines the error of not found label on the resource
// or the specified label is not found.
type ErrLabelNotFound struct {
	ErrLabelBase
}

// ErrLabelConflict defines the error of label conflicts on the resource.
type ErrLabelConflict struct {
	ErrLabelBase
}

// ErrLabelBadRequest defines the error of bad request to the resource.
type ErrLabelBadRequest struct {
	Message string
}

// NewErrLabelNotFound builds an error with ErrLabelNotFound type
func NewErrLabelNotFound(labelID int64, resourceType string, resourceIDOrName interface{}) *ErrLabelNotFound {
	return &ErrLabelNotFound{
		ErrLabelBase{
			LabelID:          labelID,
			ResourceType:     resourceType,
			ResourceIDOrName: resourceIDOrName,
		},
	}
}

// Error returns the error message of ErrLabelNotFound.
func (nf *ErrLabelNotFound) Error() string {
	if len(nf.ResourceType) > 0 && nf.ResourceIDOrName != nil {
		return fmt.Sprintf("not found: label '%d' on %s '%v'", nf.LabelID, nf.ResourceType, nf.ResourceIDOrName)
	}

	return fmt.Sprintf("not found: label '%d'", nf.LabelID)
}

// NewErrLabelConflict builds an error with NewErrLabelConflict type.
func NewErrLabelConflict(labelID int64, resourceType string, resourceIDOrName interface{}) *ErrLabelConflict {
	return &ErrLabelConflict{
		ErrLabelBase{
			LabelID:          labelID,
			ResourceType:     resourceType,
			ResourceIDOrName: resourceIDOrName,
		},
	}
}

// Error returns the error message of ErrLabelConflict.
func (cl *ErrLabelConflict) Error() string {
	return fmt.Sprintf("conflict: %s '%v' is already marked with label '%d'", cl.ResourceType, cl.ResourceIDOrName, cl.LabelID)
}

// NewErrLabelBadRequest builds an error with ErrLabelBadRequest type.
func NewErrLabelBadRequest(message string) *ErrLabelBadRequest {
	return &ErrLabelBadRequest{
		Message: message,
	}
}

// Error returns the error message of ErrLabelBadRequest.
func (br *ErrLabelBadRequest) Error() string {
	return br.Message
}
