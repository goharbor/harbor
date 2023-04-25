// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
